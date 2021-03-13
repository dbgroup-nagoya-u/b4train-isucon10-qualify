package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
)

const Limit = 20
const NazotteLimit = 50

var db *sqlx.DB
var pgConnectionData *PostgreSQLConnectionEnv
var chairSearchCondition ChairSearchCondition
var estateSearchCondition EstateSearchCondition

/*====================================================================================*
 * Type definitions
 *====================================================================================*/

type InitializeResponse struct {
	Language string `json:"language"`
}

type Chair struct {
	ID          int64  `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description"`
	Thumbnail   string `db:"thumbnail" json:"thumbnail"`
	Price       int64  `db:"price" json:"price"`
	Height      int64  `db:"height" json:"height"`
	Width       int64  `db:"width" json:"width"`
	Depth       int64  `db:"depth" json:"depth"`
	Color       string `db:"color" json:"color"`
	Features    string `db:"features" json:"features"`
	Kind        string `db:"kind" json:"kind"`
	Popularity  int64  `db:"popularity" json:"-"`
	Stock       int64  `db:"stock" json:"-"`
}

type ChairSearchResponse struct {
	Count  int64   `json:"count"`
	Chairs []Chair `json:"chairs"`
}

type ChairListResponse struct {
	Chairs []Chair `json:"chairs"`
}

//Estate 物件
type Estate struct {
	ID          int64   `db:"id" json:"id"`
	Thumbnail   string  `db:"thumbnail" json:"thumbnail"`
	Name        string  `db:"name" json:"name"`
	Description string  `db:"description" json:"description"`
	Latitude    float64 `db:"latitude" json:"latitude"`
	Longitude   float64 `db:"longitude" json:"longitude"`
	Address     string  `db:"address" json:"address"`
	Rent        int64   `db:"rent" json:"rent"`
	DoorHeight  int64   `db:"door_height" json:"doorHeight"`
	DoorWidth   int64   `db:"door_width" json:"doorWidth"`
	Features    string  `db:"features" json:"features"`
	Popularity  int64   `db:"popularity" json:"-"`
}

//EstateSearchResponse estate/searchへのレスポンスの形式
type EstateSearchResponse struct {
	Count   int64    `json:"count"`
	Estates []Estate `json:"estates"`
}

type EstateListResponse struct {
	Estates []Estate `json:"estates"`
}

type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Coordinates struct {
	Coordinates []Coordinate `json:"coordinates"`
}

type Range struct {
	ID  int64 `json:"id"`
	Min int64 `json:"min"`
	Max int64 `json:"max"`
}

type RangeCondition struct {
	Prefix string   `json:"prefix"`
	Suffix string   `json:"suffix"`
	Ranges []*Range `json:"ranges"`
}

type ListCondition struct {
	List []string `json:"list"`
}

type EstateSearchCondition struct {
	DoorWidth  RangeCondition `json:"doorWidth"`
	DoorHeight RangeCondition `json:"doorHeight"`
	Rent       RangeCondition `json:"rent"`
	Feature    ListCondition  `json:"feature"`
}

type ChairSearchCondition struct {
	Width   RangeCondition `json:"width"`
	Height  RangeCondition `json:"height"`
	Depth   RangeCondition `json:"depth"`
	Price   RangeCondition `json:"price"`
	Color   ListCondition  `json:"color"`
	Feature ListCondition  `json:"feature"`
	Kind    ListCondition  `json:"kind"`
}

type BoundingBox struct {
	// TopLeftCorner 緯度経度が共に最小値になるような点の情報を持っている
	TopLeftCorner Coordinate
	// BottomRightCorner 緯度経度が共に最大値になるような点の情報を持っている
	BottomRightCorner Coordinate
}

type PostgreSQLConnectionEnv struct {
	Host   string
	Port   string
	User   string
	DBName string
}

type RecordMapper struct {
	Record []string

	offset int
	err    error
}

/*====================================================================================*
 * Utility functions: parse record strings to values
 *====================================================================================*/

func (r *RecordMapper) next() (string, error) {
	if r.err != nil {
		return "", r.err
	}
	if r.offset >= len(r.Record) {
		r.err = fmt.Errorf("too many read")
		return "", r.err
	}
	s := r.Record[r.offset]
	r.offset++
	return s, nil
}

func (r *RecordMapper) NextInt() int {
	s, err := r.next()
	if err != nil {
		return 0
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		r.err = err
		return 0
	}
	return i
}

func (r *RecordMapper) NextFloat() float64 {
	s, err := r.next()
	if err != nil {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		r.err = err
		return 0
	}
	return f
}

func (r *RecordMapper) NextString() string {
	s, err := r.next()
	if err != nil {
		return ""
	}
	return s
}

func (r *RecordMapper) Err() error {
	return r.err
}

/*====================================================================================*
 * Utility functions: scan rows as struct
 *====================================================================================*/

func RowToChair(row *pgx.Row, chair *Chair) error {
	return (*row).Scan(
		&chair.ID,
		&chair.Name,
		&chair.Description,
		&chair.Thumbnail,
		&chair.Price,
		&chair.Height,
		&chair.Width,
		&chair.Depth,
		&chair.Color,
		&chair.Features,
		&chair.Kind,
		&chair.Popularity,
		&chair.Stock,
	)
}

func RowsToChair(rows *pgx.Rows, chair *Chair) error {
	return (*rows).Scan(
		&chair.ID,
		&chair.Name,
		&chair.Description,
		&chair.Thumbnail,
		&chair.Price,
		&chair.Height,
		&chair.Width,
		&chair.Depth,
		&chair.Color,
		&chair.Features,
		&chair.Kind,
		&chair.Popularity,
		&chair.Stock,
	)
}

func ScanChairs(rows *pgx.Rows, chairs *[]Chair) error {
	for (*rows).Next() {
		chair := Chair{}
		err := RowsToChair(rows, &chair)
		if err != nil {
			return err
		}
		*chairs = append(*chairs, chair)
	}
	return (*rows).Err()
}

func RowToEstate(row *pgx.Row, estate *Estate) error {
	return (*row).Scan(
		&estate.ID,
		&estate.Name,
		&estate.Description,
		&estate.Thumbnail,
		&estate.Address,
		&estate.Latitude,
		&estate.Longitude,
		&estate.Rent,
		&estate.DoorHeight,
		&estate.DoorWidth,
		&estate.Features,
		&estate.Popularity,
	)
}

func RowsToEstate(rows *pgx.Rows, estate *Estate) error {
	return (*rows).Scan(
		&estate.ID,
		&estate.Name,
		&estate.Description,
		&estate.Thumbnail,
		&estate.Address,
		&estate.Latitude,
		&estate.Longitude,
		&estate.Rent,
		&estate.DoorHeight,
		&estate.DoorWidth,
		&estate.Features,
		&estate.Popularity,
	)
}

func ScanEstates(rows *pgx.Rows, estates *[]Estate) error {
	for (*rows).Next() {
		estate := Estate{}
		err := RowsToEstate(rows, &estate)
		if err != nil {
			return err
		}
		*estates = append(*estates, estate)
	}
	return (*rows).Err()
}

/*====================================================================================*
 * Utility functions: getters
 *====================================================================================*/

func getRange(cond RangeCondition, rangeID string) (*Range, error) {
	RangeIndex, err := strconv.Atoi(rangeID)
	if err != nil {
		return nil, err
	}

	if RangeIndex < 0 || len(cond.Ranges) <= RangeIndex {
		return nil, fmt.Errorf("Unexpected Range ID")
	}

	return cond.Ranges[RangeIndex], nil
}

func getEnv(key, defaultValue string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}
	return defaultValue
}

/*====================================================================================*
 * Utility functions: DB connection
 *====================================================================================*/

func NewPostgreSQLConnectionEnv() *PostgreSQLConnectionEnv {
	return &PostgreSQLConnectionEnv{
		Host:   getEnv("DB_HOST", "localhost"),
		Port:   getEnv("PGPORT", "5432"),
		User:   getEnv("PGUSER", "isucon"),
		DBName: getEnv("DB_NAME", "isuumo"),
	}
}

//ConnectDB isuumoデータベースに接続する
func (env *PostgreSQLConnectionEnv) ConnectDB() (*sqlx.DB, error) {
	dsn := fmt.Sprintf("postgres://%v@%v:%v/%v", env.User, env.Host, env.Port, env.DBName)
	return sqlx.Open("pgx", dsn)
}

/*====================================================================================*
 * Utility functions: others
 *====================================================================================*/

func (cs Coordinates) coordinatesToText() string {
	points := make([]string, 0, len(cs.Coordinates))
	for _, c := range cs.Coordinates {
		points = append(points, fmt.Sprintf("%f %f", c.Latitude, c.Longitude))
	}
	return fmt.Sprintf("'POLYGON((%s))'", strings.Join(points, ","))
}

func init() {
	jsonText, err := ioutil.ReadFile("../fixture/chair_condition.json")
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	json.Unmarshal(jsonText, &chairSearchCondition)

	jsonText, err = ioutil.ReadFile("../fixture/estate_condition.json")
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	json.Unmarshal(jsonText, &estateSearchCondition)
}

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

	pgConnectionData = NewPostgreSQLConnectionEnv()

	var err error
	db, err = pgConnectionData.ConnectDB()
	if err != nil {
		e.Logger.Fatalf("DB connection failed : %v", err)
	}
	db.SetMaxOpenConns(10)
	defer db.Close()

	// Start server
	serverPort := fmt.Sprintf(":%v", getEnv("SERVER_PORT", "1323"))
	e.Logger.Fatal(e.Start(serverPort))
}

/*====================================================================================*
 * Initialization API
 *====================================================================================*/

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

func getChairDetail(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Echo().Logger.Errorf("Request parameter \"id\" parse error : %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	chair := Chair{}
	const query = `
SELECT
  *
FROM
  isuumo.chair
WHERE
  id = $1
`
	err = db.Get(&chair, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.Echo().Logger.Infof("requested id's chair not found : %v", id)
			return c.NoContent(http.StatusNotFound)
		}
		c.Echo().Logger.Errorf("Failed to get the chair from id : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	} else if chair.Stock <= 0 {
		c.Echo().Logger.Infof("requested id's chair is sold out : %v", id)
		return c.NoContent(http.StatusNotFound)
	}

	return c.JSON(http.StatusOK, chair)
}

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

	tx, err := db.Begin()
	if err != nil {
		c.Logger().Errorf("failed to begin tx: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer tx.Rollback()
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
		_, err := tx.Exec(`
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
	if err := tx.Commit(); err != nil {
		c.Logger().Errorf("failed to commit tx: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusCreated)
}

func searchChairs(c echo.Context) error {
	conditions := make([]string, 0)
	params := make([]interface{}, 0)

	if c.QueryParam("priceRangeId") != "" {
		chairPrice, err := getRange(chairSearchCondition.Price, c.QueryParam("priceRangeId"))
		if err != nil {
			c.Echo().Logger.Infof("priceRangeID invalid, %v : %v", c.QueryParam("priceRangeId"), err)
			return c.NoContent(http.StatusBadRequest)
		}

		if chairPrice.Min != -1 {
			conditions = append(conditions, "price >= $"+strconv.Itoa(len(params)+1))
			params = append(params, chairPrice.Min)
		}
		if chairPrice.Max != -1 {
			conditions = append(conditions, "price < $"+strconv.Itoa(len(params)+1))
			params = append(params, chairPrice.Max)
		}
	}

	if c.QueryParam("heightRangeId") != "" {
		chairHeight, err := getRange(chairSearchCondition.Height, c.QueryParam("heightRangeId"))
		if err != nil {
			c.Echo().Logger.Infof("heightRangeIf invalid, %v : %v", c.QueryParam("heightRangeId"), err)
			return c.NoContent(http.StatusBadRequest)
		}

		if chairHeight.Min != -1 {
			conditions = append(conditions, "height >= $"+strconv.Itoa(len(params)+1))
			params = append(params, chairHeight.Min)
		}
		if chairHeight.Max != -1 {
			conditions = append(conditions, "height < $"+strconv.Itoa(len(params)+1))
			params = append(params, chairHeight.Max)
		}
	}

	if c.QueryParam("widthRangeId") != "" {
		chairWidth, err := getRange(chairSearchCondition.Width, c.QueryParam("widthRangeId"))
		if err != nil {
			c.Echo().Logger.Infof("widthRangeID invalid, %v : %v", c.QueryParam("widthRangeId"), err)
			return c.NoContent(http.StatusBadRequest)
		}

		if chairWidth.Min != -1 {
			conditions = append(conditions, "width >= $"+strconv.Itoa(len(params)+1))
			params = append(params, chairWidth.Min)
		}
		if chairWidth.Max != -1 {
			conditions = append(conditions, "width < $"+strconv.Itoa(len(params)+1))
			params = append(params, chairWidth.Max)
		}
	}

	if c.QueryParam("depthRangeId") != "" {
		chairDepth, err := getRange(chairSearchCondition.Depth, c.QueryParam("depthRangeId"))
		if err != nil {
			c.Echo().Logger.Infof("depthRangeId invalid, %v : %v", c.QueryParam("depthRangeId"), err)
			return c.NoContent(http.StatusBadRequest)
		}

		if chairDepth.Min != -1 {
			conditions = append(conditions, "depth >= $"+strconv.Itoa(len(params)+1))
			params = append(params, chairDepth.Min)
		}
		if chairDepth.Max != -1 {
			conditions = append(conditions, "depth < $"+strconv.Itoa(len(params)+1))
			params = append(params, chairDepth.Max)
		}
	}

	if c.QueryParam("kind") != "" {
		conditions = append(conditions, "kind = $"+strconv.Itoa(len(params)+1))
		params = append(params, c.QueryParam("kind"))
	}

	if c.QueryParam("color") != "" {
		conditions = append(conditions, "color = $"+strconv.Itoa(len(params)+1))
		params = append(params, c.QueryParam("color"))
	}

	if c.QueryParam("features") != "" {
		for _, f := range strings.Split(c.QueryParam("features"), ",") {
			conditions = append(conditions, "features LIKE CONCAT('%', '"+f+"', '%')")
		}
	}

	if len(conditions) == 0 {
		c.Echo().Logger.Infof("Search condition not found")
		return c.NoContent(http.StatusBadRequest)
	}

	conditions = append(conditions, "stock > 0")

	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil {
		c.Logger().Infof("Invalid format page parameter : %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	perPage, err := strconv.Atoi(c.QueryParam("perPage"))
	if err != nil {
		c.Logger().Infof("Invalid format perPage parameter : %v", err)
		return c.NoContent(http.StatusBadRequest)
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
	searchCondition := strings.Join(conditions, "\n  AND ")
	limitOffset := `
ORDER BY
  popularity DESC,
  id ASC
LIMIT $` + strconv.Itoa(len(params)+1) + `
OFFSET $` + strconv.Itoa(len(params)+2) + `
`

	var res ChairSearchResponse
	err = db.Get(&res.Count, countQuery+searchCondition, params...)
	if err != nil {
		c.Logger().Errorf("searchChairs DB execution error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	chairs := []Chair{}
	params = append(params, perPage, page*perPage)
	err = db.Select(&chairs, searchQuery+searchCondition+limitOffset, params...)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusOK, ChairSearchResponse{Count: 0, Chairs: []Chair{}})
		}
		c.Logger().Errorf("searchChairs DB execution error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	res.Chairs = chairs

	return c.JSON(http.StatusOK, res)
}

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

	tx, err := db.Beginx()
	if err != nil {
		c.Echo().Logger.Errorf("failed to create transaction : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer tx.Rollback()

	var chair Chair
	err = tx.QueryRowx(`
SELECT
  *
FROM
  isuumo.chair
WHERE
  id = $1
  AND stock > 0
FOR UPDATE
`, id).StructScan(&chair)
	if err != nil {
		if err == sql.ErrNoRows {
			c.Echo().Logger.Infof("buyChair chair id \"%v\" not found", id)
			return c.NoContent(http.StatusNotFound)
		}
		c.Echo().Logger.Errorf("DB Execution Error: on getting a chair by id : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	_, err = tx.Exec(`
UPDATE
  isuumo.chair
SET
  stock = stock - 1
WHERE
  id = $1
`, id)
	if err != nil {
		c.Echo().Logger.Errorf("chair stock update failed : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	err = tx.Commit()
	if err != nil {
		c.Echo().Logger.Errorf("transaction commit error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

func getChairSearchCondition(c echo.Context) error {
	return c.JSON(http.StatusOK, chairSearchCondition)
}

func getLowPricedChair(c echo.Context) error {
	var chairs []Chair
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
	err := db.Select(&chairs, query, Limit)
	if err != nil {
		if err == sql.ErrNoRows {
			c.Logger().Error("getLowPricedChair not found")
			return c.JSON(http.StatusOK, ChairListResponse{[]Chair{}})
		}
		c.Logger().Errorf("getLowPricedChair DB execution error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, ChairListResponse{Chairs: chairs})
}

/*====================================================================================*
 * Estate APIs
 *====================================================================================*/

func getEstateDetail(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Echo().Logger.Infof("Request parameter \"id\" parse error : %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	var estate Estate
	err = db.Get(&estate, `
SELECT
  *
FROM
  isuumo.estate
WHERE
  id = $1
`, id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.Echo().Logger.Infof("getEstateDetail estate id %v not found", id)
			return c.NoContent(http.StatusNotFound)
		}
		c.Echo().Logger.Errorf("Database Execution error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, estate)
}

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

	tx, err := db.Begin()
	if err != nil {
		c.Logger().Errorf("failed to begin tx: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer tx.Rollback()
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
		_, err := tx.Exec(`
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
)`, id, name, description, thumbnail, address, latitude, longitude, rent, doorHeight, doorWidth, features, popularity)
		if err != nil {
			c.Logger().Errorf("failed to insert estate: %v", err)
			return c.NoContent(http.StatusInternalServerError)
		}
	}
	if err := tx.Commit(); err != nil {
		c.Logger().Errorf("failed to commit tx: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusCreated)
}

func searchEstates(c echo.Context) error {
	conditions := make([]string, 0)
	params := make([]interface{}, 0)

	if c.QueryParam("doorHeightRangeId") != "" {
		doorHeight, err := getRange(estateSearchCondition.DoorHeight, c.QueryParam("doorHeightRangeId"))
		if err != nil {
			c.Echo().Logger.Infof("doorHeightRangeID invalid, %v : %v", c.QueryParam("doorHeightRangeId"), err)
			return c.NoContent(http.StatusBadRequest)
		}

		if doorHeight.Min != -1 {
			conditions = append(conditions, "door_height >= $"+strconv.Itoa(len(params)+1))
			params = append(params, doorHeight.Min)
		}
		if doorHeight.Max != -1 {
			conditions = append(conditions, "door_height < $"+strconv.Itoa(len(params)+1))
			params = append(params, doorHeight.Max)
		}
	}

	if c.QueryParam("doorWidthRangeId") != "" {
		doorWidth, err := getRange(estateSearchCondition.DoorWidth, c.QueryParam("doorWidthRangeId"))
		if err != nil {
			c.Echo().Logger.Infof("doorWidthRangeID invalid, %v : %v", c.QueryParam("doorWidthRangeId"), err)
			return c.NoContent(http.StatusBadRequest)
		}

		if doorWidth.Min != -1 {
			conditions = append(conditions, "door_width >= $"+strconv.Itoa(len(params)+1))
			params = append(params, doorWidth.Min)
		}
		if doorWidth.Max != -1 {
			conditions = append(conditions, "door_width < $"+strconv.Itoa(len(params)+1))
			params = append(params, doorWidth.Max)
		}
	}

	if c.QueryParam("rentRangeId") != "" {
		estateRent, err := getRange(estateSearchCondition.Rent, c.QueryParam("rentRangeId"))
		if err != nil {
			c.Echo().Logger.Infof("rentRangeID invalid, %v : %v", c.QueryParam("rentRangeId"), err)
			return c.NoContent(http.StatusBadRequest)
		}

		if estateRent.Min != -1 {
			conditions = append(conditions, "rent >= $"+strconv.Itoa(len(params)+1))
			params = append(params, estateRent.Min)
		}
		if estateRent.Max != -1 {
			conditions = append(conditions, "rent < $"+strconv.Itoa(len(params)+1))
			params = append(params, estateRent.Max)
		}
	}

	if c.QueryParam("features") != "" {
		for _, f := range strings.Split(c.QueryParam("features"), ",") {
			conditions = append(conditions, "features like concat('%', '"+f+"', '%')")
		}
	}

	if len(conditions) == 0 {
		c.Echo().Logger.Infof("searchEstates search condition not found")
		return c.NoContent(http.StatusBadRequest)
	}

	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil {
		c.Logger().Infof("Invalid format page parameter : %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	perPage, err := strconv.Atoi(c.QueryParam("perPage"))
	if err != nil {
		c.Logger().Infof("Invalid format perPage parameter : %v", err)
		return c.NoContent(http.StatusBadRequest)
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
	err = db.Get(&res.Count, countQuery+searchCondition, params...)
	if err != nil {
		c.Logger().Errorf("searchEstates DB execution error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	estates := []Estate{}
	params = append(params, perPage, page*perPage)
	err = db.Select(&estates, searchQuery+searchCondition+limitOffset, params...)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusOK, EstateSearchResponse{Count: 0, Estates: []Estate{}})
		}
		c.Logger().Errorf("searchEstates DB execution error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	res.Estates = estates

	return c.JSON(http.StatusOK, res)
}

func getLowPricedEstate(c echo.Context) error {
	estates := make([]Estate, 0, Limit)
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
	err := db.Select(&estates, query, Limit)
	if err != nil {
		if err == sql.ErrNoRows {
			c.Logger().Error("getLowPricedEstate not found")
			return c.JSON(http.StatusOK, EstateListResponse{[]Estate{}})
		}
		c.Logger().Errorf("getLowPricedEstate DB execution error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, EstateListResponse{Estates: estates})
}

func searchRecommendedEstateWithChair(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Logger().Infof("Invalid format searchRecommendedEstateWithChair id : %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	chair := Chair{}
	const chairQuery = `
SELECT
  *
FROM
  isuumo.chair
WHERE
  id = $1
`
	err = db.Get(&chair, chairQuery, id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.Logger().Infof("Requested chair id \"%v\" not found", id)
			return c.NoContent(http.StatusBadRequest)
		}
		c.Logger().Errorf("Database execution error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	var estates []Estate
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
	err = db.Select(&estates, estateQuery, w, h, w, d, h, w, h, d, d, w, d, h, Limit)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusOK, EstateListResponse{[]Estate{}})
		}
		c.Logger().Errorf("Database execution error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, EstateListResponse{Estates: estates})
}

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
	estatesInBoundingBox := []Estate{}
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
	err = db.Select(&estatesInBoundingBox, extractQuery, b.BottomRightCorner.Latitude, b.TopLeftCorner.Latitude, b.BottomRightCorner.Longitude, b.TopLeftCorner.Longitude)
	if err == sql.ErrNoRows {
		c.Echo().Logger.Infof("select * from estate where latitude ...", err)
		return c.JSON(http.StatusOK, EstateSearchResponse{Count: 0, Estates: []Estate{}})
	} else if err != nil {
		c.Echo().Logger.Errorf("database execution error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	estatesInPolygon := []Estate{}
	for _, estate := range estatesInBoundingBox {
		validatedEstate := Estate{}

		point := fmt.Sprintf("'POINT(%f %f)'", estate.Latitude, estate.Longitude)
		const filterQuery = `
SELECT
  *
FROM
  isuumo.estate
WHERE
  id = $1
  AND ST_Contains(ST_PolygonFromText($2), ST_GeomFromText($3))
`
		err = db.Get(&validatedEstate, filterQuery, estate.ID, coordinates.coordinatesToText(), point)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			} else {
				c.Echo().Logger.Errorf("db access is failed on executing validate if estate is in polygon : %v", err)
				return c.NoContent(http.StatusInternalServerError)
			}
		} else {
			estatesInPolygon = append(estatesInPolygon, validatedEstate)
		}
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

	estate := Estate{}
	const query = `
SELECT
  *
FROM
  isuumo.estate
WHERE
  id = $1
`
	err = db.Get(&estate, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.NoContent(http.StatusNotFound)
		}
		c.Logger().Errorf("postEstateRequestDocument DB execution error : %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

func getEstateSearchCondition(c echo.Context) error {
	return c.JSON(http.StatusOK, estateSearchCondition)
}

func (cs Coordinates) getBoundingBox() BoundingBox {
	coordinates := cs.Coordinates
	boundingBox := BoundingBox{
		TopLeftCorner: Coordinate{
			Latitude: coordinates[0].Latitude, Longitude: coordinates[0].Longitude,
		},
		BottomRightCorner: Coordinate{
			Latitude: coordinates[0].Latitude, Longitude: coordinates[0].Longitude,
		},
	}
	for _, coordinate := range coordinates {
		if boundingBox.TopLeftCorner.Latitude > coordinate.Latitude {
			boundingBox.TopLeftCorner.Latitude = coordinate.Latitude
		}
		if boundingBox.TopLeftCorner.Longitude > coordinate.Longitude {
			boundingBox.TopLeftCorner.Longitude = coordinate.Longitude
		}

		if boundingBox.BottomRightCorner.Latitude < coordinate.Latitude {
			boundingBox.BottomRightCorner.Latitude = coordinate.Latitude
		}
		if boundingBox.BottomRightCorner.Longitude < coordinate.Longitude {
			boundingBox.BottomRightCorner.Longitude = coordinate.Longitude
		}
	}
	return boundingBox
}
