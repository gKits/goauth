package db

import "time"

/*
connectToRedis
*/
func connectToRedis() {}

/*
BlacklistJWT takes a tokenString of a jwt and
a lifetime and will add it to the Redis key-value
store to blacklist it for the duration of its
lifetime
If the blacklisting fails it will return an error
*/
func BlacklistJWT(tokenString string, lifetime time.Duration) error {
    return nil
}

/*
IsBlacklisted returns true when the given tokenString
is found in the Redis false if it's not
*/
func IsBlacklisted(tokenString string) (bool, error) {
    return false, nil
}
