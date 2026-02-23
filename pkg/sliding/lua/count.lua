-- Sliding window counter: atomically removes expired entries, adds current
-- event, sets TTL, and returns the window count.
-- KEYS[1] = window key
-- ARGV[1] = min score (epoch ms, window start)
-- ARGV[2] = max score (epoch ms, now)
-- ARGV[3] = current score (epoch ms, now — used as member score)
-- ARGV[4] = TTL in seconds

local key   = KEYS[1]
local min   = ARGV[1]
local now   = ARGV[2]
local ttl   = tonumber(ARGV[4])

redis.call('ZREMRANGEBYSCORE', key, '-inf', min)
redis.call('ZADD', key, now, now)
redis.call('EXPIRE', key, ttl)
return redis.call('ZCOUNT', key, min, '+inf')
