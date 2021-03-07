all: chair_data estate_data

refresh_chair_data:
	find ../webapp/frontend/public/images/chair -name *.png | xargs rm

refresh_estate_data:
	find ../webapp/frontend/public/images/estate -name *.png | xargs rm

chair_data: make_chair_data.py
	mkdir -p result/draft_data/chair
	python3 make_chair_data.py
	cp result/2_DummyChairData.sql ../webapp/mysql/db/2_DummyChairData.sql
	cp result/chair_condition.json ../webapp/fixture/chair_condition.json

estate_data: make_estate_data.py
	mkdir -p result/draft_data/estate
	python3 make_estate_data.py
	cp result/1_DummyEstateData.sql ../webapp/mysql/db/1_DummyEstateData.sql
	cp result/estate_condition.json ../webapp/fixture/estate_condition.json
