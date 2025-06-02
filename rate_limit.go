package main

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/turbot/pipe-fittings/v2/error_helpers"
	"github.com/turbot/pipe-fittings/v2/parse"
	pfplugin "github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"strconv"
	"strings"
)

func setRateLimiters(limiterConfigStr string) error {

	limiters, err := parseLimiterConfig(limiterConfigStr)
	if err != nil {
		return err
	}
	// for now we only use the limiter config
	if len(limiters) == 0 {
		return nil
	}
	// build a SetRateLimitersRequest
	req := &proto.SetRateLimitersRequest{
		Definitions: make([]*proto.RateLimiterDefinition, len(limiters)),
	}
	for i, l := range limiters {
		req.Definitions[i] = RateLimiterAsProto(l)
	}

	// set the plugin config on the plugin server
	_, err = pluginServer.SetRateLimiters(req)
	return err
}

func parseLimiterConfig(configString string) ([]*pfplugin.RateLimiter, error) {
	parser := hclparse.NewParser()
	cfg, err := unescapeConfig(configString)
	if err != nil {
		return nil, fmt.Errorf("failed to unescape plugin config: %w", err)
	}
	file, diags := parser.ParseHCL([]byte(cfg), "input.hcl")
	if diags.HasErrors() {
		return nil, error_helpers.HclDiagsToError("failed to parse plugin config", diags)
	}

	content, _, contentDiags := file.Body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "limiter", // Replace with your block type
				LabelNames: []string{"name"},
			},
		},
	})
	if contentDiags.HasErrors() {
		return nil, contentDiags
	}

	if len(content.Blocks) == 0 {
		return nil, hcl.Diagnostics{}
	}

	var limiters []*pfplugin.RateLimiter
	for _, block := range content.Blocks {
		if block.Type != "limiter" {
			continue // Skip blocks that are not of type "limiter"
		}

		l, moreDiags := parse.DecodeLimiter(block)
		if moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
			continue // Skip this block if there are errors
		}
		limiters = append(limiters, l)
	}
	if diags.HasErrors() {
		return nil, error_helpers.HclDiagsToError("failed to parse plugin config", diags)
	}
	return limiters, nil
}

func unescapeConfig(s string) (string, error) {
	// Wrap in quotes and use strconv.Unquote to unescape
	return strconv.Unquote(`"` + strings.ReplaceAll(s, `"`, `\"`) + `"`)
}

func RateLimiterAsProto(l *pfplugin.RateLimiter) *proto.RateLimiterDefinition {
	res := &proto.RateLimiterDefinition{
		Name:  l.Name,
		Scope: l.Scope,
	}
	if l.MaxConcurrency != nil {
		res.MaxConcurrency = *l.MaxConcurrency
	}
	if l.BucketSize != nil {
		res.BucketSize = *l.BucketSize
	}
	if l.FillRate != nil {
		res.FillRate = *l.FillRate
	}
	if l.Where != nil {
		res.Where = *l.Where
	}

	return res
}

