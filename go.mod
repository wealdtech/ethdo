module github.com/wealdtech/ethdo

go 1.21

toolchain go1.21.6

require (
	github.com/attestantio/go-eth2-client v0.21.4
	github.com/ferranbt/fastssz v0.1.3
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/google/uuid v1.6.0
	github.com/hako/durafmt v0.0.0-20210608085754-5c1018a4e16b
	github.com/herumi/bls-eth-go-binary v1.35.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/nbutton23/zxcvbn-go v0.0.0-20210217022336-fa2cb2858354
	github.com/pkg/errors v0.9.1
	github.com/prysmaticlabs/go-bitfield v0.0.0-20240328144219-a1caa50c3a1e
	github.com/prysmaticlabs/go-ssz v0.0.0-20210121151755-f6208871c388
	github.com/rs/zerolog v1.33.0
	github.com/shopspring/decimal v1.4.0
	github.com/spf13/cobra v1.8.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.18.2
	github.com/stretchr/testify v1.9.0
	github.com/tyler-smith/go-bip39 v1.1.0
	github.com/wealdtech/go-bytesutil v1.2.1
	github.com/wealdtech/go-ecodec v1.1.4
	github.com/wealdtech/go-eth2-types/v2 v2.8.2
	github.com/wealdtech/go-eth2-util v1.8.2
	github.com/wealdtech/go-eth2-wallet v1.16.0
	github.com/wealdtech/go-eth2-wallet-dirk v1.4.9
	github.com/wealdtech/go-eth2-wallet-distributed v1.2.1
	github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4 v1.4.1
	github.com/wealdtech/go-eth2-wallet-hd/v2 v2.7.0
	github.com/wealdtech/go-eth2-wallet-nd/v2 v2.5.0
	github.com/wealdtech/go-eth2-wallet-store-filesystem v1.18.1
	github.com/wealdtech/go-eth2-wallet-store-s3 v1.12.0
	github.com/wealdtech/go-eth2-wallet-store-scratch v1.7.2
	github.com/wealdtech/go-eth2-wallet-types/v2 v2.11.0
	github.com/wealdtech/go-string2eth v1.2.1
	golang.org/x/text v0.16.0
)

require (
	github.com/aws/aws-sdk-go v1.53.10 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgraph-io/ristretto v0.1.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fatih/color v1.17.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/goccy/go-yaml v1.11.3 // indirect
	github.com/golang/glog v1.2.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/holiman/uint256 v1.2.4 // indirect
	github.com/huandu/go-clone v1.7.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/pk910/dynamic-ssz v0.0.4 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.19.1 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.53.0 // indirect
	github.com/prometheus/procfs v0.15.0 // indirect
	github.com/protolambda/zssz v0.1.5 // indirect
	github.com/r3labs/sse/v2 v2.10.0 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/shibukawa/configdir v0.0.0-20170330084843-e180dbdc8da0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/wealdtech/eth2-signer-api v1.7.2 // indirect
	github.com/wealdtech/go-indexer v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.52.0 // indirect
	go.opentelemetry.io/otel v1.27.0 // indirect
	go.opentelemetry.io/otel/metric v1.27.0 // indirect
	go.opentelemetry.io/otel/trace v1.27.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.24.0 // indirect
	golang.org/x/exp v0.0.0-20240525044651-4c93da0ed11d // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/xerrors v0.0.0-20231012003039-104605ab7028 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240521202816-d264139d666e // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240521202816-d264139d666e // indirect
	google.golang.org/grpc v1.64.1 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
	gopkg.in/Knetic/govaluate.v3 v3.0.0 // indirect
	gopkg.in/cenkalti/backoff.v1 v1.1.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
