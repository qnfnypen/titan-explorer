-- name: GetDeviceInfo :one
SELECT * FROM `device_info` WHERE device_id = ? LIMIT 1;



