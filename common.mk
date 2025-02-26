# This contains Makefile logic that is common to several makefiles

BUILD_TAGS ?= pelldvs

COMMIT_HASH := $(shell git rev-parse --short HEAD)
LD_FLAGS = -X github.com/0xPellNetwork/pelldvs/version.TMGitCommitHash=$(COMMIT_HASH)
BUILD_FLAGS = -mod=readonly -ldflags "$(LD_FLAGS)"
# allow users to pass additional flags via the conventional LDFLAGS variable
LD_FLAGS += $(LDFLAGS)

# handle nostrip
ifeq (,$(findstring nostrip,$(PELLDVS_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
  LD_FLAGS += -s -w
endif

# handle race
ifeq (race,$(findstring race,$(PELLDVS_BUILD_OPTIONS)))
  CGO_ENABLED=1
  BUILD_FLAGS += -race
endif

# handle cleveldb
ifeq (cleveldb,$(findstring cleveldb,$(PELLDVS_BUILD_OPTIONS)))
  CGO_ENABLED=1
  BUILD_TAGS += cleveldb
endif

# handle badgerdb
ifeq (badgerdb,$(findstring badgerdb,$(PELLDVS_BUILD_OPTIONS)))
  BUILD_TAGS += badgerdb
endif

# handle rocksdb
ifeq (rocksdb,$(findstring rocksdb,$(PELLDVS_BUILD_OPTIONS)))
  CGO_ENABLED=1
  BUILD_TAGS += rocksdb
endif

# handle boltdb
ifeq (boltdb,$(findstring boltdb,$(PELLDVS_BUILD_OPTIONS)))
  BUILD_TAGS += boltdb
endif
