module github.com/affordmed/vehicle_maintenance_scheduler

go 1.25.4

replace github.com/affordmed/logging_middleware => ../logging_middleware

require github.com/affordmed/logging_middleware v0.0.0-00010101000000-000000000000

require (
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/ilyakaznacheev/cleanenv v1.5.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	olympos.io/encoding/edn v0.0.0-20201019073823-d3554ca0b0a3 // indirect
)
