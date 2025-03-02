package config

import (
	"fmt"
	"regexp"
	"time"

	"github.com/Trendyol/chaki/util/slc"
	"github.com/Trendyol/chaki/util/wrapper"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

// Wrapper is config wrapper method, it can be used to manipulate config on injection process
type Wrapper wrapper.Wrapper[*Config]

// Config is a higher-level Viper wrapper that comes with some extra features
type Config struct {
	references map[string]*viper.Viper
	prefix     string
	v          *viper.Viper
}

// NewConfig construct a config from viper instances
func NewConfig(v *viper.Viper, references map[string]*viper.Viper) *Config {
	r := map[string]*viper.Viper{
		"this": v,
		"env":  newEnvViper(),
	}

	for key, value := range references {
		r[key] = value
	}

	c := &Config{
		references: r,
		prefix:     "",
		v:          v,
	}

	c.parseReferences()

	return c
}

func newEnvViper() *viper.Viper {
	v := viper.New()
	v.AutomaticEnv()
	return v
}

// NewConfigFromPaths construct a config from paths
func NewConfigFromPaths(path string, referencePaths map[string]string) (*Config, error) {
	v, err := readConfig(path)
	if err != nil {
		return nil, err
	}

	references := make(map[string]*viper.Viper)

	for k, p := range referencePaths {
		refv, err := readConfig(p)
		if err != nil {
			return nil, err
		}
		references[k] = refv
	}

	return NewConfig(v, references), nil
}

func (g *Config) GetBool(k string) bool {
	return getValueT(g.v, g.key(k), cast.ToBoolE)
}

func (g *Config) GetDuration(k string) time.Duration {
	return getValueT(g.v, g.key(k), cast.ToDurationE)
}

func (g *Config) GetFloat64(k string) float64 {
	return getValueT(g.v, g.key(k), cast.ToFloat64E)
}

func (g *Config) GetInt(k string) int {
	return getValueT(g.v, g.key(k), cast.ToIntE)
}

func (g *Config) GetInt32(k string) int32 {
	return getValueT(g.v, g.key(k), cast.ToInt32E)
}

func (g *Config) GetInt64(k string) int64 {
	return getValueT(g.v, g.key(k), cast.ToInt64E)
}

func (g *Config) GetIntSlice(k string) []int {
	return getValueT(g.v, g.key(k), cast.ToIntSliceE)
}

func (g *Config) GetString(k string) string {
	return getValueT(g.v, g.key(k), cast.ToStringE)
}

func (g *Config) GetStringMap(k string) map[string]any {
	return getValueT(g.v, g.key(k), cast.ToStringMapE)
}

func (g *Config) GetStringSlice(k string) []string {
	return getValueT(g.v, g.key(k), cast.ToStringSliceE)
}

func (g *Config) GetTime(k string) time.Time {
	return getValueT(g.v, g.key(k), cast.ToTimeE)
}

func (g *Config) SetDefault(key string, value any) {
	g.v.SetDefault(g.key(key), value)
}

func (g *Config) Get(key string) any {
	return getValue(g.v, g.key(key))
}

func (g *Config) Set(key string, value any) {
	g.v.Set(g.key(key), value)
}

func ToStruct[T any](cfg *Config, key string) (t T, err error) {
	err = cfg.v.UnmarshalKey(cfg.key(key), &t)
	return
}

func (g *Config) parseReferences() {
	slc.ForEach(g.v.AllKeys(), func(key string) {
		vv := g.v.Get(key)
		if refkey, value, ok := parseReferenceValue(vv); ok {
			ref := g.references[refkey]
			if ref == nil {
				panic("no reference found for: " + refkey)
			}
			vv = getValueByReference(ref, value, g.references)
		}
		g.Set(key, vv)
	})
}

// Exists check a key exists or not
func (g *Config) Exists(key string) bool {
	return g.v.Get(g.key(key)) != nil
}

// Of returns a new config instance with provided key prefix
// Example:
//
//	cfg.GetString("foo.bar") == cfg.Of("foo").GetString("bar") // true
func (g *Config) Of(prefix string) *Config {
	if g.prefix != "" {
		prefix = fmt.Sprintf("%s.%s", g.prefix, prefix)
	}
	return &Config{
		v:          g.v,
		prefix:     prefix,
		references: g.references,
	}
}

func (g *Config) key(k string) string {
	if g.prefix != "" {
		return fmt.Sprintf("%s.%s", g.prefix, k)
	}
	return k
}

func getValueT[T any](v *viper.Viper, key string, caster func(any) (T, error)) T {
	t := getValue(v, key)
	r, err := caster(t)
	if err != nil {
		panic(err.Error())
	}
	return r
}

func getValue(v *viper.Viper, key string) any {
	t := v.Get(key)
	if t == nil {
		panic(fmt.Sprintf("key not found in config %s", key))
	}
	return t
}

func getValueByReference(v *viper.Viper, key string, references map[string]*viper.Viper) any {
	t := v.Get(key)
	if t == nil {
		panic(fmt.Sprintf("key not found in config %s", key))
	}
	key, val, ok := parseReferenceValue(t)
	if ok {
		ref := references[key]
		if ref == nil {
			panic("no reference found for: " + key)
		}
		return getValueByReference(ref, val, references)
	}
	return t
}

var referenceKeyRegexp = regexp.MustCompile(`\$\{([A-Za-z0-9\-_]*):([A-Za-z0-9\-_\.]*)\}`)

func parseReferenceValue(v any) (key string, ref string, ok bool) {
	s, ok := v.(string)
	if !ok {
		return "", "", false
	}

	if !referenceKeyRegexp.MatchString(s) {
		return "", "", false
	}

	return referenceKeyRegexp.ReplaceAllString(s, "$1"), referenceKeyRegexp.ReplaceAllString(s, "$2"), true
}
