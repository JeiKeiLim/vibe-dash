// Package config provides configuration loading and management using Viper.
// It implements the ports.ConfigLoader interface for YAML-based configuration.
//
// YAML Configuration Format:
//
// The config file at ~/.vibe-dash/config.yaml uses the following structure:
//
//	settings:
//	  hibernation_days: 14
//	  refresh_interval_seconds: 10
//	  refresh_debounce_ms: 200
//	  agent_waiting_threshold_minutes: 10
//
//	projects:
//	  project-id:
//	    path: /path/to/project
//	    display_name: My Project
//	    favorite: true
//	    hibernation_days: 7  # optional override
//	    agent_waiting_threshold_minutes: 5  # optional override
//
// Viper handles the YAML parsing directly using its map-based approach,
// which is then mapped to ports.Config via mapViperToConfig().
package config
