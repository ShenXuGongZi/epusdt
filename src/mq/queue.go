package mq

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/assimon/luuu/mq/handle"
	"github.com/assimon/luuu/util/log"
	"github.com/hibiken/asynq"
	"github.com/spf13/viper"
)

var MClient *asynq.Client

func buildRedisOpt() asynq.RedisClientOpt {
	// If redis_url is set (e.g. rediss://default:PASS@HOST:PORT for Upstash),
	// parse it to extract host/password/db/TLS.
	if rawURL := viper.GetString("redis_url"); rawURL != "" {
		u, err := url.Parse(rawURL)
		if err == nil {
			opt := asynq.RedisClientOpt{
				Addr: u.Host,
			}
			if pwd, ok := u.User.Password(); ok {
				opt.Password = pwd
			}
			if u.User.Username() != "" {
				opt.Username = u.User.Username()
			}
			// DB from path (e.g., /5)
			if len(u.Path) > 1 {
				if db, err := strconv.Atoi(strings.TrimPrefix(u.Path, "/")); err == nil {
					opt.DB = db
				}
			}
			// Enable TLS for rediss:// scheme
			if u.Scheme == "rediss" {
				opt.TLSConfig = &tls.Config{ServerName: u.Hostname()}
			}
			return opt
		}
		log.Sugar.Errorf("[queue] redis_url parse failed: %v, falling back to host/port", err)
	}
	return asynq.RedisClientOpt{
		Addr: fmt.Sprintf(
			"%s:%s",
			viper.GetString("redis_host"),
			viper.GetString("redis_port")),
		DB:       viper.GetInt("redis_db"),
		Password: viper.GetString("redis_passwd"),
	}
}

func Start() {
	redisOpt := buildRedisOpt()
	initClient(redisOpt)
	go initListen(redisOpt)
}

func initClient(redis asynq.RedisClientOpt) {
	MClient = asynq.NewClient(redis)
}

func initListen(redis asynq.RedisClientOpt) {
	srv := asynq.NewServer(
		redis,
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: viper.GetInt("queue_concurrency"),
			// Optionally specify multiple queues with different priority.
			Queues: map[string]int{
				"critical": viper.GetInt("queue_level_critical"),
				"default":  viper.GetInt("queue_level_default"),
				"low":      viper.GetInt("queue_level_low"),
			},
			Logger: log.Sugar,
		},
	)
	mux := asynq.NewServeMux()
	mux.HandleFunc(handle.QueueOrderExpiration, handle.OrderExpirationHandle)
	mux.HandleFunc(handle.QueueOrderCallback, handle.OrderCallbackHandle)
	if err := srv.Run(mux); err != nil {
		log.Sugar.Fatalf("[queue] could not run server: %v", err)
	}
}
