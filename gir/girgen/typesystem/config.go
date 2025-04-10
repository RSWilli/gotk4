package typesystem

import (
	"fmt"
	"log/slog"
	"maps"

	"github.com/diamondburned/gotk4/gir"
)

type Config struct {
	// GIRReplacements allows you to rename a GIR type. This may be needed for primitives that are implemented in another namespace.
	//
	// make sure that these replacements terminate or the typesystem will crash
	GIRReplacements map[string]string

	// Namespaces contains the configuration for the given versioned namespace (e.g. GLib-2). If a key is missing in this map,
	// then the typesystem will resolve everything in that namespace
	Namespaces map[string]NamespaceConfig
}

type NamespaceConfig struct {
	// Ignored signifies that the whole namespace should be treated as not existing
	Ignored bool

	// MinVersion declares the minimal version that should be supported in
	// type resolution. Everything that is deprecated longer than this version will be ignored.
	MinVersion string

	IgnoredDefinitions []IgnoreFunc

	// ManualTypes contains the gir name to a manual type override that will not be generated. Themanual type
	// must be in the same go package as the generator would place it.
	ManualTypes []Type
}

func (cfg Config) getNamespaceEnv(namespace *Namespace) *env {
	nsCfg := cfg.Namespaces[fmt.Sprintf("%s-%d", namespace.Name, namespace.Version.Major)]

	if nsCfg.Ignored {
		return nil
	}

	var err error
	var v gir.Version

	if nsCfg.MinVersion != "" {
		v, err = gir.ParseVersion(nsCfg.MinVersion)

		if err != nil {
			panic(err)
		}
	}

	return &env{
		cfg:        cfg,
		nsCfg:      nsCfg,
		minVersion: v,
		ignore:     ignoreOr(nsCfg.IgnoredDefinitions...),
		namespace:  namespace,
		logger:     slog.Default().With(slog.String("namespace", namespace.v.String())),
	}
}

// Combine combines both configs, needed for extensions of the base configs. It assumes that this can be done without
// collosions
func (cfg Config) Combine(other Config) Config {
	var newCfg Config

	if len(cfg.GIRReplacements) > 0 || len(other.GIRReplacements) > 0 {
		newCfg.GIRReplacements = make(map[string]string)
	}

	maps.Copy(newCfg.GIRReplacements, cfg.GIRReplacements)
	maps.Copy(newCfg.GIRReplacements, other.GIRReplacements)

	if len(cfg.Namespaces) > 0 || len(other.Namespaces) > 0 {
		newCfg.Namespaces = make(map[string]NamespaceConfig)
	}

	maps.Copy(newCfg.Namespaces, cfg.Namespaces)
	maps.Copy(newCfg.Namespaces, other.Namespaces)

	return newCfg
}
