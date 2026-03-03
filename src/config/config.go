package config

import (
	"fmt"
	"github.com/spf13/viper"
	"net/url"
	"os"
	"time"
)

var (
	AppDebug    bool
	MysqlDns    string
	RuntimePath string
	LogSavePath string
	StaticPath  string
	TgBotToken  string
	TgProxy     string
	TgManage    int64
	UsdtRate    float64
)

func Init() {
	// Read from .env file if it exists (optional for Railway/Docker deployments)
	viper.AddConfigPath("./")
	viper.SetConfigFile(".env")
	_ = viper.ReadInConfig() // ignore error if .env file not found

	// Also read from environment variables (Railway passes config as env vars)
	// Bind each key explicitly so lowercase env vars work (viper.AutomaticEnv uppercases by default)
	for _, key := range []string{
		"app_name", "app_uri", "app_debug", "http_listen",
		"static_path", "runtime_root_path", "log_save_path", "log_max_size", "log_max_age", "max_backups",
		"mysql_host", "mysql_port", "mysql_user", "mysql_passwd", "mysql_database",
		"mysql_table_prefix", "mysql_max_idle_conns", "mysql_max_open_conns", "mysql_max_life_time",
		"redis_host", "redis_port", "redis_passwd", "redis_db", "redis_pool_size", "redis_max_retries", "redis_idle_timeout",
		"queue_concurrency", "queue_level_critical", "queue_level_default", "queue_level_low",
		"tg_bot_token", "tg_proxy", "tg_manage",
		"api_auth_token", "order_expiration_time", "forced_usdt_rate",
	} {
		viper.BindEnv(key, key)
	}
	viper.AutomaticEnv()

	gwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	AppDebug = viper.GetBool("app_debug")
	StaticPath = viper.GetString("static_path")
	RuntimePath = fmt.Sprintf(
		"%s%s",
		gwd,
		viper.GetString("runtime_root_path"))
	LogSavePath = fmt.Sprintf(
		"%s%s",
		RuntimePath,
		viper.GetString("log_save_path"))
	// Build MySQL DSN; append tls=true if connecting to a remote host (e.g. PlanetScale)
	mysqlHost := viper.GetString("mysql_host")
	mysqlAddr := fmt.Sprintf("%s:%s", mysqlHost, viper.GetString("mysql_port"))
	tlsParam := ""
	if mysqlHost != "127.0.0.1" && mysqlHost != "localhost" {
		tlsParam = "&tls=true"
	}
	MysqlDns = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local%s",
		url.QueryEscape(viper.GetString("mysql_user")),
		url.QueryEscape(viper.GetString("mysql_passwd")),
		mysqlAddr,
		viper.GetString("mysql_database"),
		tlsParam)
	TgBotToken = viper.GetString("tg_bot_token")
	TgProxy = viper.GetString("tg_proxy")
	TgManage = viper.GetInt64("tg_manage")
}

func GetAppVersion() string {
	return "0.0.2"
}

func GetAppName() string {
	appName := viper.GetString("app_name")
	if appName == "" {
		return "epusdt"
	}
	return appName
}

func GetAppUri() string {
	return viper.GetString("app_uri")
}

func GetApiAuthToken() string {
	return viper.GetString("api_auth_token")
}

func GetUsdtRate() float64 {
	forcedUsdtRate := viper.GetFloat64("forced_usdt_rate")
	if forcedUsdtRate > 0 {
		return forcedUsdtRate
	}
	if UsdtRate <= 0 {
		return 6.4
	}
	return UsdtRate
}

func GetOrderExpirationTime() int {
	timer := viper.GetInt("order_expiration_time")
	if timer <= 0 {
		return 10
	}
	return timer
}

func GetOrderExpirationTimeDuration() time.Duration {
	timer := GetOrderExpirationTime()
	return time.Minute * time.Duration(timer)
}
