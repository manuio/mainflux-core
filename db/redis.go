/**
 * Copyright (c) Mainflux
 *
 * Mainflux server is licensed under an Apache license, version 2.0.
 * All rights not explicitly granted in the Apache license, version 2.0 are reserved.
 * See the included LICENSE file for more details.
 */

package db

import (
	"github.com/garyburd/redigo/redis"
	"strconv"
)

const maxIdle int = 10

var pool *redis.Pool

// Redis struct
type Redis struct {
	Conn redis.Conn
}

// Start starts new redis pool with allowed maximum of 10 inactive connections
func InitRedis(host string, port int) error {
	var err error

	if mainSession == nil {
		pool = &redis.Pool{
			MaxIdle: maxIdle,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", host+":"+strconv.Itoa(port))
				if err != nil {
					return nil, err
				}

				return c, err
			},
		}
	}

	return err
}

// Init function retrieves and sets a redis connection from the connection pool
func (r *Redis) Init() error {
	var err error
	r.Conn = pool.Get()
	return err
}

// Publish publishes message
func (r *Redis) Publish(c string, msg string) error {
	_, err := r.Conn.Do("PUBLISH", c, msg)
	return err
}

// Close closes current connection
func (r *Redis) Close() error {
	return r.Conn.Close()
}

// Stop terminates the redis pool.
func StopRedis() error {
	return pool.Close()
}
