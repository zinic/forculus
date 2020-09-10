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

func compileS3Uploaders(cfg configuration) (map[string]S3Uploader, error) {
	uploaders := make(map[string]S3Uploader, len(cfg.S3Uploaders))
	for name, rawS3Uploader := range cfg.S3Uploaders {
		if filter, err := parseAlertFilterFields(rawS3Uploader.Filter); err != nil {
			return nil, fmt.Errorf("s3 uploader %s contains a fatal configuration err: %w", name, err)
		} else {
			uploaders[name] = S3Uploader{
				Filter:     filter,
				s3Uploader: rawS3Uploader,
			}
		}
	}

	return uploaders, nil
}

func compileEmailAlerts(cfg configuration) (map[string]EmailAlert, error) {
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

func validateEmailAlertSMTPReferences(alerts map[string]EmailAlert, cfg configuration) error {
	for _, compiledAlert := range alerts {
		if _, serverValid := cfg.SMTPServers[compiledAlert.Server]; !serverValid {
			return fmt.Errorf("email alert %s references an unknown SMTP server %s", compiledAlert.Name, compiledAlert.Server)
		}
	}

	return nil
}

func parseConfigurationFields(cfg configuration) (Configuration, error) {
	compiledCfg := Configuration{
		Zoneminder: cfg.Zoneminder,
		SMTPServer: cfg.SMTPServers,
	}

	if compiledS3Uploaders, err := compileS3Uploaders(cfg); err != nil {
		return compiledCfg, err
	} else {
		compiledCfg.S3Uploaders = compiledS3Uploaders
	}

	if compiledAlerts, err := compileEmailAlerts(cfg); err != nil {
		return compiledCfg, err
	} else if err := validateEmailAlertSMTPReferences(compiledAlerts, cfg); err != nil {
		return compiledCfg, err
	} else {
		compiledCfg.EmailAlerts = compiledAlerts
	}

	return compiledCfg, nil
}
