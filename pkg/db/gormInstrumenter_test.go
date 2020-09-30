package db

import (
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type TestType struct {
	gorm.Model
	Code  string
	Price uint
}

func migrateFunc(conn *gorm.DB) error {
	return conn.AutoMigrate(&TestType{}).Error
}

func TestGormInstrumenter(t *testing.T) {
	InitializeTestSqlite3(migrateFunc)
	defer Close(db)

	metricSrv := httptest.NewServer(promhttp.Handler())
	defer metricSrv.Close()

	registerInstrumenterPlugin()

	var testType TestType

	t.Run("Create metric collected", func(t *testing.T) {
		db.Create(&TestType{Code: "L1212", Price: 1000})

		checkMetricsCollection(metricSrv, []string{
			`d4l_db_request_duration_seconds_bucket{sqlstring="INSERT INTO \"test_types\" (\"cr",le="0.25"} 1`,
		}, t)

	})
	t.Run("Query metric collected", func(t *testing.T) {
		db.First(&testType, 1)

		checkMetricsCollection(metricSrv, []string{
			`d4l_db_request_duration_seconds_bucket{sqlstring="SELECT * FROM \"test_types\"  W",le="0.25"} 1`,
		}, t)

	})

	t.Run("Update metric collected", func(t *testing.T) {
		db.Model(&testType).Update("Price", 2000)

		checkMetricsCollection(metricSrv, []string{
			`d4l_db_request_duration_seconds_bucket{sqlstring="UPDATE \"test_types\" SET \"pric",le="0.25"} 1`,
		}, t)

	})

	t.Run("Delete metric collected", func(t *testing.T) {
		db.Delete(&testType)

		checkMetricsCollection(metricSrv, []string{
			`d4l_db_request_duration_seconds_bucket{sqlstring="UPDATE \"test_types\" SET \"pric",le="0.25"} 1`,
		}, t)

	})

}

func BenchmarkWithMetrics(b *testing.B) {
	InitializeTestSqlite3(migrateFunc)
	defer Close(db)

	var testType TestType
	db.Create(&TestType{Code: "L1212", Price: 1000})

	for i := 0; i < b.N; i++ {
		db.First(&testType, 1)
	}
}

func checkMetricsCollection(metricSrv *httptest.Server, want []string, t *testing.T) {
	resp, err := metricSrv.Client().Get(metricSrv.URL)
	if err != nil {
		t.Fatalf("error getting metrics: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading body from metrics response: %v", err)
	}
	bodyStr := string(body)

	for _, contain := range want {
		if !strings.Contains(bodyStr, contain) {
			t.Fatalf("%s is not collected: %s", contain, bodyStr)
		}
	}
}
