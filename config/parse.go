package config

import (
	"fmt"
	"regexp"
)

func parseAlertFilterFields(cfg alertFilter) (AlertFilter, error) {
	filter := AlertFilter{
		EventTrigger:        cfg.EventTrigger,
		AlertFrameThreshold: cfg.AlertFrameThreshold,
	}

	if len(cfg.NameFilterRegex) > 0 {
		if compiledRegex, err := regexp.Compile(cfg.NameFilterRegex); err != nil {
			return filter, fmt.Errorf("alert filter name regex is malformed: %w", err)
		} else {
			filter.NameRegex = compiledRegex
		}
	}

	return filter, nil
}

func compileUploaders(cfg eventServerConfiguration) (map[string]Uploader, error) {
	uploaders := make(map[string]Uploader, len(cfg.Uploaders))
	for name, rawUploader := range cfg.Uploaders {
		if filter, err := parseAlertFilterFields(rawUploader.Filter); err != nil {
			return nil, fmt.Errorf("uploader %s has a malformed configuration: %w", name, err)
		} else {
			uploaders[name] = Uploader{
				StorageTarget: rawUploader.StorageTarget,
				Filter:        filter,
			}
		}
	}

	return uploaders, nil
}

func compileEmailAlerts(cfg eventServerConfiguration) (map[string]EmailAlert, error) {
	alerts := make(map[string]EmailAlert, len(cfg.EmailAlerts))
	for name, rawEmailAlert := range cfg.EmailAlerts {
		if filter, err := parseAlertFilterFields(rawEmailAlert.Filter); err != nil {
			return nil, fmt.Errorf("s3 uploader %s contains a fatal configuration err: %w", name, err)
		} else {
			alerts[name] = EmailAlert{
				Filter:     filter,
				emailAlert: rawEmailAlert,
			}
		}
	}

	return alerts, nil
}

func validateStorageProviderReferences(cfg EventServerConfig) error {
	for uploaderName, uploaderCfg := range cfg.Uploaders {
		if _, providerExists := cfg.StorageProviders[uploaderCfg.StorageTarget]; !providerExists {
			return fmt.Errorf("uploader %s references an unknown storage provider %s", uploaderName, uploaderCfg.StorageTarget)
		}
	}

	return nil
}

func validateEmailAlertSMTPReferences(cfg EventServerConfig) error {
	for alertName, alertCfg := range cfg.EmailAlerts {
		if _, serverExists := cfg.SMTPServers[alertCfg.Server]; !serverExists {
			return fmt.Errorf("email alert %s references an unknown SMTP server %s", alertName, alertCfg.Server)
		}
	}

	return nil
}

func parseEventServerCfg(cfg eventServerConfiguration) (EventServerConfig, error) {
	compiledCfg := EventServerConfig{
		Zoneminder:       cfg.Zoneminder,
		StorageProviders: cfg.StorageProviders,
		SMTPServers:      cfg.SMTPServers,
	}

	if compiledUploaders, err := compileUploaders(cfg); err != nil {
		return compiledCfg, err
	} else {
		compiledCfg.Uploaders = compiledUploaders
	}

	if compiledAlerts, err := compileEmailAlerts(cfg); err != nil {
		return compiledCfg, err
	} else {
		compiledCfg.EmailAlerts = compiledAlerts
	}

	if err := validateEmailAlertSMTPReferences(compiledCfg); err != nil {
		return compiledCfg, err
	}

	if err := validateStorageProviderReferences(compiledCfg); err != nil {
		return compiledCfg, err
	}

	return compiledCfg, nil
}
