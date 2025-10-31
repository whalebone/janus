module github.com/hellofresh/janus

go 1.25

require (
	code.cloudfoundry.org/bytefmt v0.56.0
	contrib.go.opencensus.io/exporter/jaeger v0.2.1
	contrib.go.opencensus.io/exporter/prometheus v0.4.2
	github.com/Knetic/govaluate v3.0.1-0.20171022003610-9aa49832a739+incompatible
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/asaskevich/govalidator v0.0.0-20171111151018-521b25f4b05f
	github.com/cucumber/godog v0.12.6
	github.com/cucumber/messages-go/v16 v16.0.1
	github.com/felixge/httpsnoop v1.0.4
	github.com/fsnotify/fsnotify v1.9.0
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/gocql/gocql v1.7.0
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/uuid v1.6.0
	github.com/hellofresh/health-go/v3 v3.2.0
	github.com/hellofresh/logging-go v0.4.0
	github.com/hellofresh/opencensus-go-extras v0.0.0-20201019073047-da70cdf01446
	github.com/hellofresh/stats-go v0.8.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/magiconair/properties v1.8.10
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.5.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/rafaeljesus/retry-go v0.0.0-20171214204623-5981a380a879
	github.com/redis/go-redis/v9 v9.16.0
	github.com/rs/cors v1.11.1
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.10.1
	github.com/spf13/viper v1.21.0
	github.com/stretchr/testify v1.11.1
	github.com/tidwall/gjson v1.18.0
	github.com/ulule/limiter/v3 v3.11.2
	go.mongodb.org/mongo-driver v1.17.6
	go.opencensus.io v0.24.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.63.0
	go.opentelemetry.io/otel v1.38.0
	go.opentelemetry.io/otel/bridge/opencensus v1.38.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.38.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.38.0
	go.opentelemetry.io/otel/sdk v1.38.0
	golang.org/x/crypto v0.43.0
	golang.org/x/net v0.46.0
	golang.org/x/oauth2 v0.32.0
)

require (
	github.com/DataDog/datadog-go v4.8.3+incompatible // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/TV4/logrus-stackdriver-formatter v0.1.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bshuster-repo/logrus-logstash-hook v0.4.1 // indirect
	github.com/cactus/go-statsd-client v3.1.1+incompatible // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cucumber/gherkin-go/v19 v19.0.3 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fiam/gounidecode v0.0.0-20150629112515-8deddbd03fec // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-memdb v1.3.5 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/klauspost/compress v1.18.1 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.2 // indirect
	github.com/prometheus/procfs v0.19.2 // indirect
	github.com/prometheus/statsd_exporter v0.28.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20250401214520-65e299d6c5c9 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/smartystreets/goconvey v1.8.1 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/uber/jaeger-client-go v2.25.0+incompatible // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.38.0 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	go.opentelemetry.io/proto/otlp v1.8.0 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/api v0.254.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251029180050-ab9386a59fda // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	google.golang.org/grpc v1.76.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/alexcesaro/statsd.v2 v2.0.0 // indirect
	gopkg.in/gemnasium/logrus-graylog-hook.v2 v2.0.7 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
