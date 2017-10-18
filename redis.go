package main

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
	"github.com/jhunt/vcaptive"

	"github.com/starkandwayne/shield-cf-demo/internal/rand"
)

const RedisVerificationKey = "redis-verification-key"
const RedisDataKey = "dat"

type RedisSystem struct {
	redis redis.Conn
}

func (s *RedisSystem) Configure(services vcaptive.Services) (bool, error) {
	svc, found := services.Tagged("redis")
	if !found {
		return false, nil
	}

	host, ok := svc.GetString("host")
	if !ok {
		return true, fmt.Errorf("VCAP_SERVICES: '%s' service has no 'host' credential", svc.Label)
	}

	port, ok := svc.GetUint("port")
	if !ok {
		return true, fmt.Errorf("VCAP_SERVICES: '%s' service has no 'port' credential", svc.Label)
	}

	password, ok := svc.GetString("password")
	if !ok {
		return true, fmt.Errorf("VCAP_SERVICES: '%s' service has no 'password' credential", svc.Label)
	}

	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", host, port), redis.DialPassword(password))
	if err != nil {
		return true, err
	}

	s.redis = conn
	return true, nil
}

func (s *RedisSystem) Setup() error {
	if exists, _ := redis.Bool(s.redis.Do("EXISTS", RedisVerificationKey)); exists {
		return nil
	}

	if _, err := s.redis.Do("SET", RedisVerificationKey, rand.VerificationKey()); err != nil {
		return err
	}

	for i := 0; i < rand.Bound(2048, 128); i++ {
		if _, err := s.redis.Do("RPUSH", RedisDataKey, fmt.Sprintf("%x", i)); err != nil {
			return err
		}
	}

	return nil
}

func (s *RedisSystem) Teardown() error {
	if _, err := s.redis.Do("DEL", RedisDataKey); err != nil {
		return err
	}

	if _, err := s.redis.Do("DEL", RedisVerificationKey); err != nil {
		return err
	}

	return nil
}

func (s *RedisSystem) Summarize() Data {
	dat := Data{
		System:       "Redis",
		Summary:      "              *no data found*\n",
		Verification: "UNKNOWN",
		OK:           false,
	}

	if exists, _ := redis.Bool(s.redis.Do("EXISTS", RedisVerificationKey)); !exists {
		return dat
	}

	n, err := redis.Int(s.redis.Do("LLEN", RedisDataKey))
	if err != nil {
		dat.Summary = fmt.Sprintf("ERROR:        *%s*\n", err)
		return dat
	}

	vfy, err := redis.String(s.redis.Do("GET", RedisVerificationKey))
	if err != nil {
		dat.Summary = fmt.Sprintf("ERROR:        *%s*\n", err)
		return dat
	}

	dat.Verification = vfy
	dat.Summary = fmt.Sprintf("Stored Keys:  *%d*\n", n)
	dat.OK = true
	return dat
}

func init() {
	if Systems == nil {
		Systems = make(map[string]System)
	}
	Systems["redis"] = &RedisSystem{}
}
