package tinygraph

// Most of the options are delegated to RocksDB.

// We forked
// https://github.com/DanielMorsing/rocksdb/blob/master/options.go in
// order to expose more RocksDB options.

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	rocks "github.csv.comcast.com/jsteph206/gorocksdb"
)

const (
	DefaultCacheSize = 1 << 20
)

type Options map[string]interface{}

// Filename can either be a file name or JSON.  Surprise!
func LoadOptions(filename string) (*Options, error) {
	var bs []byte
	if strings.HasPrefix(filename, "{") {
		bs = []byte(filename)
	} else {
		var err error
		log.Printf("LoadOptions reading %s", filename)
		bs, err = ioutil.ReadFile(filename)
		if err != nil {
			log.Printf("LoadOptions error: %v (%s)", err, filename)
			return nil, err
		}
	}

	opts := make(Options)
	err := json.Unmarshal(bs, &opts)
	if err != nil {
		log.Printf("LoadOptions unmarshal error: %v on \"%s\"", err, bs)
		return nil, err
	}

	return &opts, nil
}

func (d Options) IntKey(key string) (int, bool) {
	if val, ok := d[key]; ok {
		switch vv := val.(type) {
		case float64:
			return int(vv), true
		default:
			panic(fmt.Errorf("Invalid '%s' parameter type from config."))
		}
	}
	return 0, false
}

func (d Options) StringKey(key string) (string, bool) {
	if val, ok := d[key]; ok {
		switch vv := val.(type) {
		case string:
			return vv, true
		default:
			panic(fmt.Errorf("Invalid '%s' parameter type from config."))
		}
	}
	return "", false
}

func (d Options) BoolKey(key string) (bool, bool) {
	if val, ok := d[key]; ok {
		switch vv := val.(type) {
		case bool:
			return vv, true
		default:
			panic(fmt.Errorf("Invalid '%s' parameter type from config."))
		}
	}
	return false, false
}

// https://github.com/facebook/rocksdb/blob/master/include/rocksdb/c.h
// https://github.com/facebook/rocksdb/blob/master/include/rocksdb/options.h

func RocksOpts(options *Options) *rocks.Options {

	if options == nil {
		m := make(Options)
		options = &m
	}

	// ToDo
	env := rocks.NewDefaultEnv()

	if n, ok := options.IntKey("background_threads"); ok {
		env.SetBackgroundThreads(n)
		log.Printf("config env.SetBackgroundThreads(%d)\n", n)
	}

	if n, ok := options.IntKey("high_priority_background_threads"); ok {
		env.SetHighPriorityBackgroundThreads(n)
		log.Printf("config env.SetHighPriorityBackgroundThreads(%d)\n", n)
	}

	opts := rocks.NewOptions()
	opts.SetEnv(env)

	if b, ok := options.BoolKey("read_only"); ok {
		opts.SetReadOnly(b)
		log.Printf("config opts.SetReadOnly(%v)\n", b)
	}

	cacheSize := DefaultCacheSize
	if n, ok := options.IntKey("cache_size"); ok {
		cacheSize = n
	}
	cache := rocks.NewLRUCache(cacheSize)
	opts.SetCache(cache)

	// opts.SetComparator(cmp)
	opts.SetErrorIfExists(false)

	if n, ok := options.IntKey("increase_parallelism"); ok {
		opts.IncreaseParallelism(n)
		log.Printf("config opts.IncreaseParallelism(%d)\n", n)
	}

	if b, ok := options.BoolKey("disable_data_sync"); ok {
		opts.SetDisableDataSync(b)
		log.Printf("config opts.SetDisableDataSync(%v)\n", b)
	}

	if n, ok := options.IntKey("bytes_per_sync_power"); ok {
		opts.SetBytesPerSync(uint64(1) << uint64(n))
		log.Printf("config opts.SetBytesPerSync(%d)\n", uint64(1)<<uint64(n))
	}

	if n, ok := options.IntKey("log_level"); ok {
		opts.SetLogLevel(n)
		log.Printf("config opts.SetLogLevel(%d)\n", n)
	}

	if dir, ok := options.StringKey("log_dir"); ok {
		opts.SetLogDir(dir)
		log.Printf("config opts.SetLogDir(%s)\n", dir)
	}

	if dir, ok := options.StringKey("wal_dir"); ok {
		opts.SetLogDir(dir)
		log.Printf("config opts.SetWalDir(%s)\n", dir)
	}

	if n, ok := options.IntKey("stats_dump_period"); ok {
		opts.SetStatsDumpPeriod(uint(n))
		log.Printf("config opts.SetStatsDumpPeriod(%d)\n", n)
	}

	if n, ok := options.IntKey("write_buffer_size"); ok {
		opts.SetWriteBufferSize(n)
		log.Printf("config opts.SetWriteBufferSize(%d)\n", n)
	}

	if n, ok := options.IntKey("write_buffer_size_power"); ok {
		opts.SetWriteBufferSize(int(uint64(1) << uint64(n)))
		log.Printf("config opts.SetWriteBufferSize(%d)\n", int(uint64(1)<<uint(n)))
	}

	if b, ok := options.BoolKey("paranoid_checks"); ok {
		opts.SetParanoidChecks(b)
		log.Printf("config opts.SetParanoidChecks(%v)\n", b)
	}

	if b, ok := options.BoolKey("allow_mmap_reads"); ok {
		opts.SetAllowMMapReads(b)
		log.Printf("config opts.SetAllowMMapReads(%v)\n", b)
	}

	if b, ok := options.BoolKey("allow_mmap_writes"); ok {
		opts.SetAllowMMapWrites(b)
		log.Printf("config opts.SetAllowMMapWrites(%v)\n", b)
	}

	if b, ok := options.BoolKey("allow_os_buffer"); ok {
		opts.SetAllowOSBuffer(b)
		log.Printf("config opts.SetAllowOSBuffer(%v)\n", b)
	}

	if n, ok := options.IntKey("max_open_files"); ok {
		opts.SetMaxOpenFiles(n)
		log.Printf("config opts.SetMaxOpenFiles(%d)\n", n)
	}

	if n, ok := options.IntKey("max_write_buffer_number"); ok {
		opts.SetMaxWriteBufferNumber(n)
		log.Printf("config opts.SetMaxWriteBufferNumber(%d)\n", n)
	}

	if n, ok := options.IntKey("min_write_buffer_number_to_merge"); ok {
		opts.SetMinWriteBufferNumberToMerge(n)
		log.Printf("config opts.SetMinWriteBufferNumberToMerge(%d)\n", n)
	}

	if n, ok := options.IntKey("block_size"); ok {
		opts.SetBlockSize(n)
		log.Printf("config opts.SetBlockSize(%d)\n", n)
	}

	if n, ok := options.IntKey("block_restart_interval"); ok {
		opts.SetBlockRestartInterval(n)
		log.Printf("config opts.SetBlockRestartInterval(%d)\n", n)
	}

	// Compaction

	if n, ok := options.IntKey("num_levels"); ok {
		opts.SetNumLevels(n)
		log.Printf("config opts.SetNumLevels(%d)\n", n)
	}

	if n, ok := options.IntKey("level0_num_file_compaction_trigger"); ok {
		opts.SetLevel0FileNumCompactionTrigger(n)
		log.Printf("config opts.SetLevel0FileNumCompactionTrigger(%d)\n", n)
	}

	if n, ok := options.IntKey("target_file_size_base_power"); ok {
		opts.SetTargetFileSizeBase(uint64(1) << uint64(n))
		log.Printf("config opts.SetTargetFileSizeBase(%d)\n", uint64(1)<<uint64(n))
	}

	if n, ok := options.IntKey("target_file_size_multiplier"); ok {
		opts.SetTargetFileSizeMultiplier(n)
		log.Printf("config opts.SetTargetFileSizeMultiplier(%d)\n", n)
	}

	if n, ok := options.IntKey("max_background_compactions"); ok {
		opts.SetMaxBackgroundCompactions(n)
		log.Printf("config opts.SetMaxBackgroundCompactions(%d)\n", n)
	}

	if n, ok := options.IntKey("max_background_flushes"); ok {
		opts.SetMaxBackgroundFlushes(n)
		log.Printf("config opts.SetMaxBackgroundFlushes(%d)\n", n)
	}

	comp, ok := options.StringKey("compression")
	if !ok {
		opts.SetCompression(rocks.NoCompression)
	} else {
		// ToDo: https://github.com/facebook/rocksdb/blob/master/include/rocksdb/c.h#L520-L527
		switch comp {
		case "snappy":
			opts.SetCompression(rocks.SnappyCompression)
		case "none":
			opts.SetCompression(rocks.NoCompression)
		default:
			panic(fmt.Errorf("Bad compression: %s", comp))
			return nil
		}
	}

	opts.SetCreateIfMissing(true)
	opts.SetErrorIfExists(false)

	return opts
}

func RocksReadOpts(options *Options) *rocks.ReadOptions {
	// ToDo
	if options == nil {
		m := make(Options)
		options = &m
	}

	opts := rocks.NewReadOptions()

	if b, ok := options.BoolKey("verify_checksums"); ok {
		opts.SetVerifyChecksums(b)
		log.Printf("config opts.SetVerifyChecksums(%v)\n", b)
	}
	if b, ok := options.BoolKey("fill_cache_size"); ok {
		opts.SetFillCache(b)
		log.Printf("config opts.SetFillCache(%v)\n", b)
	}

	return opts
}

func RocksWriteOpts(options *Options) *rocks.WriteOptions {
	// ToDo
	if options == nil {
		m := make(Options)
		options = &m
	}

	opts := rocks.NewWriteOptions()

	if b, ok := options.BoolKey("sync"); ok {
		opts.SetSync(b)
		log.Printf("config opts.SetSync(%v)\n", b)
	}

	if b, ok := options.BoolKey("disable_wal"); ok {
		opts.DisableWAL(b)
		log.Printf("config opts.DisableWAL(%v)\n", b)
	}
	return opts
}
