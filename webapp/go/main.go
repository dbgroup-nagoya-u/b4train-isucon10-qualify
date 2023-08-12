package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const Limit = 20
const NazotteLimit = 50

var ctx context.Context
var pool *pgxpool.Pool
var pgConnectionData *DBConnEnv
var chairSearchCondition ChairSearchCondition
var estateSearchCondition EstateSearchCondition

/*====================================================================================*
 * Main function
 *====================================================================================*/

func main() {
	// Echo instance
	e := echo.New()
	e.Debug = true
	e.Logger.SetLevel(log.DEBUG)

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Initialize
	e.POST("/initialize", initialize)

	// Chair Handler
	e.GET("/api/chair/:id", getChairDetail)
	e.POST("/api/chair", postChair)
	e.GET("/api/chair/search", searchChairs)
	e.GET("/api/chair/low_priced", getLowPricedChair)
	e.GET("/api/chair/search/condition", getChairSearchCondition)
	e.POST("/api/chair/buy/:id", buyChair)

	// Estate Handler
	e.GET("/api/estate/:id", getEstateDetail)
	e.POST("/api/estate", postEstate)
	e.GET("/api/estate/search", searchEstates)
	e.GET("/api/estate/low_priced", getLowPricedEstate)
	e.POST("/api/estate/req_doc/:id", postEstateRequestDocument)
	e.POST("/api/estate/nazotte", searchEstateNazotte)
	e.GET("/api/estate/search/condition", getEstateSearchCondition)
	e.GET("/api/recommended_estate/:id", searchRecommendedEstateWithChair)

	ctx = context.Background()
	pgConnectionData = NewDBConnEnv()

	var err error
	pool, err = pgConnectionData.ConnectDB()
	if err != nil {
		e.Logger.Fatalf("DB connection failed : %v", err)
	}
	defer pool.Close()

	// Start server
	serverPort := fmt.Sprintf(":%v", getEnv("SERVER_PORT", "1323"))
	e.Logger.Fatal(e.Start(serverPort))
}

/*====================================================================================*
 * Initialization API
 *====================================================================================*/

// ベンチマーク用にデータベース内の情報を初期化する．
//
// http://<FQDN>/initialize
func initialize(c echo.Context) error {
	sqlDir := filepath.Join("..", "psql", "db")
	absPath, _ := filepath.Abs(sqlDir)
	cmdStr := fmt.Sprintf("%v/init.sh", absPath)
	if err := exec.Command("bash", "-c", cmdStr).Run(); err != nil {
		c.Logger().Errorf("Initialize script error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, InitializeResponse{
		Language: "go",
	})
}

/*====================================================================================*
 * Chair APIs
 *====================================================================================*/

// 指定したIDの椅子について詳細な情報を取得する．
//
// http://<FQDN>/api/chair/:id
func getChairDetail(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Echo().Logger.Errorf("Request parameter \"id\" parse error : %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	const query = `
SELECT
  *
FROM
  isuumo.chair
WHERE
  id = $1
`
	var chair Chair
	err = pgxscan.Get(ctx, pool, &chair, query, id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Echo().Logger.Infof("requested id's chair not found : %v", id)
			return c.NoContent(http.StatusNotFound)
		}
		c.Echo().Logger.Errorf("Failed to get the chair from id : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if chair.Stock <= 0 {
		c.Echo().Logger.Infof("requested id's chair is sold out : %v", id)
		return c.NoContent(http.StatusNotFound)
	}

	return c.JSON(http.StatusOK, chair)
}

// 新しい椅子を登録する．
//
// http://<FQDN>/api/chair
func postChair(c echo.Context) error {
	header, err := c.FormFile("chairs")
	if err != nil {
		c.Logger().Errorf("failed to get form file: %v", err)
		return c.NoContent(http.StatusBadRequest)
	}
	f, err := header.Open()
	if err != nil {
		c.Logger().Errorf("failed to open form file: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer f.Close()

	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		c.Logger().Errorf("failed to read csv: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		c.Logger().Errorf("failed to begin tx: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer tx.Rollback(ctx)

	for _, row := range records {
		rm := RecordMapper{Record: row}
		id := rm.NextInt()
		name := rm.NextString()
		description := rm.NextString()
		thumbnail := rm.NextString()
		price := rm.NextInt()
		height := rm.NextInt()
		width := rm.NextInt()
		depth := rm.NextInt()
		color := rm.NextString()
		features := rm.NextString()
		kind := rm.NextString()
		popularity := rm.NextInt()
		stock := rm.NextInt()
		if err := rm.Err(); err != nil {
			c.Logger().Errorf("failed to read record: %v", err)
			return c.NoContent(http.StatusBadRequest)
		}
		_, err := tx.Exec(
			ctx,
			`
INSERT INTO isuumo.chair (
  id,
  name,
  description,
  thumbnail,
  price,
  height,
  width,
  depth,
  color,
  features,
  kind,
  popularity,
  stock
) VALUES(
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12,
  $13
)
`,
			id, name, description, thumbnail, price, height, width, depth, color, features, kind, popularity, stock)
		if err != nil {
			c.Logger().Errorf("failed to insert chair: %v", err)
			return c.NoContent(http.StatusInternalServerError)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		c.Logger().Errorf("failed to commit tx: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusCreated)
}

// 指定した検索条件にマッチする椅子の情報を取得する．
//
// http://<FQDN>/api/chair/search
func searchChairs(c echo.Context) error {
	conditions, params, page, perPage, err := parseChairSearchConditions(c)
	if err != nil {
		return err
	}

	const searchQuery = `
SELECT
  *
FROM
  isuumo.chair
WHERE
  `
	const countQuery = `
SELECT
  COUNT(*)
FROM
  isuumo.chair
WHERE
  `
	limitOffset := `
ORDER BY
  popularity DESC,
  id ASC
LIMIT
  $` + strconv.Itoa(len(params)+1) + `
OFFSET
  $` + strconv.Itoa(len(params)+2) + `
`
	searchCondition := strings.Join(conditions, "\n  AND ")

	var res ChairSearchResponse
	query := countQuery + searchCondition
	err = pgxscan.Get(ctx, pool, &res.Count, query, params...)

	if err != nil {
		c.Logger().Errorf("searchChairs DB execution error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if res.Count <= 0 {
		return c.JSON(http.StatusOK, ChairSearchResponse{Count: 0, Chairs: []Chair{}})
	}

	chairs := make([]Chair, 0, perPage)
	query = searchQuery + searchCondition + limitOffset
	params = append(params, perPage, page*perPage)
	err = pgxscan.Select(ctx, pool, &chairs, query, params...)

	if err != nil {
		c.Logger().Errorf("searchChairs DB execution error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	res.Chairs = chairs
	return c.JSON(http.StatusOK, res)
}

// 価格の低い順に椅子の情報を取得する．
//
// http://<FQDN>/api/chair/low_priced
func getLowPricedChair(c echo.Context) error {
	const query = `
SELECT
  *
FROM
  isuumo.chair
WHERE
  stock > 0
ORDER BY
  price ASC,
  id ASC
LIMIT $1
`
	chairs := make([]Chair, 0, Limit)
	err := pgxscan.Select(ctx, pool, &chairs, query, Limit)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Logger().Error("getLowPricedChair not found")
			return c.JSON(http.StatusOK, ChairListResponse{[]Chair{}})
		}
		c.Logger().Errorf("getLowPricedChair DB execution error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, ChairListResponse{Chairs: chairs})
}

// 椅子の検索条件の一覧を取得する．
//
// http://<FQDN>/api/chair/search/condition
func getChairSearchCondition(c echo.Context) error {
	return c.JSON(http.StatusOK, chairSearchCondition)
}

// 指定したIDの椅子を1つ購入する．
//
// http://<FQDN>/api/chair/buy/:id
func buyChair(c echo.Context) error {
	m := echo.Map{}
	if err := c.Bind(&m); err != nil {
		c.Echo().Logger.Infof("post buy chair failed : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	_, ok := m["email"].(string)
	if !ok {
		c.Echo().Logger.Info("post buy chair failed : email not found in request body")
		return c.NoContent(http.StatusBadRequest)
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Echo().Logger.Infof("post buy chair failed : %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		c.Echo().Logger.Errorf("failed to create transaction : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer tx.Rollback(ctx)

	const chairQuery = `
SELECT
  *
FROM
  isuumo.chair
WHERE
  id = $1
  AND stock > 0
FOR UPDATE
`
	var chair Chair
	err = pgxscan.Get(ctx, pool, &chair, chairQuery, id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Echo().Logger.Infof("buyChair chair id \"%v\" not found", id)
			return c.NoContent(http.StatusNotFound)
		}
		c.Echo().Logger.Errorf("DB Execution Error: on getting a chair by id : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	const updateQuery = `
UPDATE
  isuumo.chair
SET
  stock = stock - 1
WHERE
  id = $1
`
	_, err = tx.Exec(ctx, updateQuery, id)
	if err != nil {
		c.Echo().Logger.Errorf("chair stock update failed : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	err = tx.Commit(ctx)
	if err != nil {
		c.Echo().Logger.Errorf("transaction commit error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

/*====================================================================================*
 * Estate APIs
 *====================================================================================*/

// 指定したIDの物件の詳細情報を取得する．
//
// http://<FQDN>/api/estate/:id
func getEstateDetail(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Echo().Logger.Infof("Request parameter \"id\" parse error : %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	const query = `
SELECT
  *
FROM
  isuumo.estate
WHERE
  id = $1
`
	var estate Estate
	err = pgxscan.Get(ctx, pool, &estate, query, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Echo().Logger.Infof("getEstateDetail estate id %v not found", id)
			return c.NoContent(http.StatusNotFound)
		}
		c.Echo().Logger.Errorf("Database Execution error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, estate)
}

// 新しい物件の情報を登録する．
//
// http://<FQDN>/api/estate
func postEstate(c echo.Context) error {
	header, err := c.FormFile("estates")
	if err != nil {
		c.Logger().Errorf("failed to get form file: %v", err)
		return c.NoContent(http.StatusBadRequest)
	}
	f, err := header.Open()
	if err != nil {
		c.Logger().Errorf("failed to open form file: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer f.Close()

	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		c.Logger().Errorf("failed to read csv: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		c.Logger().Errorf("failed to begin tx: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer tx.Rollback(ctx)

	for _, row := range records {
		rm := RecordMapper{Record: row}
		id := rm.NextInt()
		name := rm.NextString()
		description := rm.NextString()
		thumbnail := rm.NextString()
		address := rm.NextString()
		latitude := rm.NextFloat()
		longitude := rm.NextFloat()
		rent := rm.NextInt()
		doorHeight := rm.NextInt()
		doorWidth := rm.NextInt()
		features := rm.NextString()
		popularity := rm.NextInt()
		if err := rm.Err(); err != nil {
			c.Logger().Errorf("failed to read record: %v", err)
			return c.NoContent(http.StatusBadRequest)
		}

		const query = `
INSERT INTO isuumo.estate (
  id,
  name,
  description,
  thumbnail,
  address,
  latitude,
  longitude,
  rent,
  door_height,
  door_width,
  features,
  popularity
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12
)`
		_, err := tx.Exec(
			ctx, query,
			id, name, description, thumbnail, address, latitude, longitude, rent, doorHeight,
			doorWidth, features, popularity,
		)
		if err != nil {
			c.Logger().Errorf("failed to insert estate: %v", err)
			return c.NoContent(http.StatusInternalServerError)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		c.Logger().Errorf("failed to commit tx: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusCreated)
}

// 指定した検索条件にマッチする物件の情報を取得する．
//
// http://<FQDN>/api/estate/search
func searchEstates(c echo.Context) error {
	conditions, params, page, perPage, err := parseEstateSearchConditions(c)
	if err != nil {
		return err
	}

	const searchQuery = `
SELECT
  *
FROM
  isuumo.estate
WHERE
  `
	const countQuery = `
SELECT
  COUNT(*)
FROM
  isuumo.estate
WHERE
`
	searchCondition := strings.Join(conditions, "\n  AND ")
	limitOffset := `
ORDER BY
  popularity DESC,
  id ASC
LIMIT $` + strconv.Itoa(len(params)+1) + `
OFFSET $` + strconv.Itoa(len(params)+2) + `
`

	var res EstateSearchResponse
	query := countQuery + searchCondition
	err = pgxscan.Get(ctx, pool, &res.Count, query, params...)

	if err != nil {
		c.Logger().Errorf("searchEstates DB execution error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if res.Count <= 0 {
		return c.JSON(http.StatusOK, EstateSearchResponse{Count: 0, Estates: []Estate{}})
	}

	estates := make([]Estate, 0, perPage)
	query = searchQuery + searchCondition + limitOffset
	params = append(params, perPage, page*perPage)
	err = pgxscan.Select(ctx, pool, &estates, query, params...)

	if err != nil {
		c.Logger().Errorf("searchEstates DB execution error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	res.Estates = estates
	return c.JSON(http.StatusOK, res)
}

// 価格の低い順に物件の情報を取得する．
//
// http://<FQDN>/api/estate/low_priced
func getLowPricedEstate(c echo.Context) error {
	const query = `
SELECT
  *
FROM
  isuumo.estate
ORDER BY
  rent ASC,
  id ASC
LIMIT $1
`
	estates := make([]Estate, 0, Limit)
	err := pgxscan.Select(ctx, pool, &estates, query, Limit)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Logger().Error("getLowPricedEstate not found")
			return c.JSON(http.StatusOK, EstateListResponse{[]Estate{}})
		}
		c.Logger().Errorf("getLowPricedEstate DB execution error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, EstateListResponse{Estates: estates})
}

// 指定したIDの物件についての資料請求を処理する．
//
// http://<FQDN>/api/estate/req_doc/:id
func postEstateRequestDocument(c echo.Context) error {
	m := echo.Map{}
	if err := c.Bind(&m); err != nil {
		c.Echo().Logger.Infof("post request document failed : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	_, ok := m["email"].(string)
	if !ok {
		c.Echo().Logger.Info("post request document failed : email not found in request body")
		return c.NoContent(http.StatusBadRequest)
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Echo().Logger.Infof("post request document failed : %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	const query = `
SELECT
  *
FROM
  isuumo.estate
WHERE
  id = $1
`
	var estate Estate
	err = pgxscan.Get(ctx, pool, &estate, query, id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.NoContent(http.StatusNotFound)
		}
		c.Logger().Errorf("postEstateRequestDocument DB execution error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// 指定された領域に存在する物件の情報を取得する．
//
// http://<FQDN>/api/estate/nazotte
func searchEstateNazotte(c echo.Context) error {
	coordinates := Coordinates{}
	err := c.Bind(&coordinates)
	if err != nil {
		c.Echo().Logger.Infof("post search estate nazotte failed : %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	if len(coordinates.Coordinates) == 0 {
		return c.NoContent(http.StatusBadRequest)
	}

	b := coordinates.getBoundingBox()

	const extractQuery = `
SELECT
  *
FROM
  isuumo.estate
WHERE
  latitude <= $1
  AND latitude >= $2
  AND longitude <= $3
  AND longitude >= $4
ORDER BY
  popularity DESC,
  id ASC
`
	estatesInBoundingBox := make([]Estate, 0, NazotteLimit)
	err = pgxscan.Select(ctx, pool, &estatesInBoundingBox, extractQuery,
		b.BottomRightCorner.Latitude, b.TopLeftCorner.Latitude,
		b.BottomRightCorner.Longitude, b.TopLeftCorner.Longitude,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Echo().Logger.Infof("select * from estate where latitude ...", err)
			return c.JSON(http.StatusOK, EstateSearchResponse{Count: 0, Estates: []Estate{}})
		}
		c.Echo().Logger.Errorf("database execution error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	estatesInPolygon := make([]Estate, 0, NazotteLimit)
	for _, estate := range estatesInBoundingBox {
		point := fmt.Sprintf("'POINT(%f %f)'", estate.Latitude, estate.Longitude)
		filterQuery := `
SELECT
  *
FROM
  isuumo.estate
WHERE
  id = $1
  AND ST_Contains(
    ST_PolygonFromText(` + coordinates.coordinatesToText() + `),
    ST_GeomFromText(` + point + `)
  )
`

		var validatedEstate Estate
		err = pgxscan.Get(ctx, pool, &validatedEstate, filterQuery, estate.ID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			c.Echo().Logger.Errorf("db access is failed on executing validate if estate is in polygon : %v", err)
			return c.NoContent(http.StatusInternalServerError)
		}

		estatesInPolygon = append(estatesInPolygon, validatedEstate)
	}

	var re EstateSearchResponse
	re.Estates = []Estate{}
	if len(estatesInPolygon) > NazotteLimit {
		re.Estates = estatesInPolygon[:NazotteLimit]
	} else {
		re.Estates = estatesInPolygon
	}
	re.Count = int64(len(re.Estates))

	return c.JSON(http.StatusOK, re)
}

// 物件の検索条件の一覧を取得する．
//
// http://<FQDN>/api/estate/search/condition
func getEstateSearchCondition(c echo.Context) error {
	return c.JSON(http.StatusOK, estateSearchCondition)
}

// 指定したIDの椅子を搬入可能な物件の情報を取得する．
//
// http://<FQDN>/api/recommended_estate/:id
func searchRecommendedEstateWithChair(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Logger().Infof("Invalid format searchRecommendedEstateWithChair id : %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	const chairQuery = `
SELECT
  *
FROM
  isuumo.chair
WHERE
  id = $1
`
	var chair Chair
	err = pgxscan.Get(ctx, pool, &chair, chairQuery, id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.Logger().Infof("Requested chair id \"%v\" not found", id)
			return c.NoContent(http.StatusBadRequest)
		}
		c.Logger().Errorf("Database execution error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	w := chair.Width
	h := chair.Height
	d := chair.Depth
	const estateQuery = `
SELECT
  *
FROM
  isuumo.estate
WHERE
  (door_width >= $1 AND door_height >= $2)
  OR (door_width >= $3 AND door_height >= $4)
  OR (door_width >= $5 AND door_height >= $6)
  OR (door_width >= $7 AND door_height >= $8)
  OR (door_width >= $9 AND door_height >= $10)
  OR (door_width >= $11 AND door_height >= $12)
ORDER BY
  popularity DESC,
  id ASC
LIMIT $13
`
	estates := make([]Estate, 0, Limit)
	err = pgxscan.Select(ctx, pool, &estates, estateQuery,
		w, h, w, d, h, w, h, d, d, w, d, h, Limit,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusOK, EstateListResponse{[]Estate{}})
		}
		c.Logger().Errorf("Database execution error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, EstateListResponse{Estates: estates})
}
