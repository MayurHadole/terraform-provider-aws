package aws

import (
	"fmt"
	"math"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func wafv2EmptySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
}

func wafv2EmptySchemaDeprecated() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
		Deprecated: "Not supported by WAFv2 API",
	}
}

func wafv2RootStatementSchema(level int) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"and_statement":                         wafv2StatementSchema(level - 1),
				"byte_match_statement":                  wafv2ByteMatchStatementSchema(),
				"geo_match_statement":                   wafv2GeoMatchStatementSchema(),
				"ip_set_reference_statement":            wafv2IpSetReferenceStatementSchema(),
				"not_statement":                         wafv2StatementSchema(level - 1),
				"or_statement":                          wafv2StatementSchema(level - 1),
				"regex_pattern_set_reference_statement": wafv2RegexPatternSetReferenceStatementSchema(),
				"size_constraint_statement":             wafv2SizeConstraintSchema(),
				"sqli_match_statement":                  wafv2SqliMatchStatementSchema(),
				"xss_match_statement":                   wafv2XssMatchStatementSchema(),
			},
		},
	}
}

func wafv2StatementSchema(level int) *schema.Schema {
	if level > 1 {
		return &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"statement": {
						Type:     schema.TypeList,
						Required: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"and_statement":                         wafv2StatementSchema(level - 1),
								"byte_match_statement":                  wafv2ByteMatchStatementSchema(),
								"geo_match_statement":                   wafv2GeoMatchStatementSchema(),
								"ip_set_reference_statement":            wafv2IpSetReferenceStatementSchema(),
								"not_statement":                         wafv2StatementSchema(level - 1),
								"or_statement":                          wafv2StatementSchema(level - 1),
								"regex_pattern_set_reference_statement": wafv2RegexPatternSetReferenceStatementSchema(),
								"size_constraint_statement":             wafv2SizeConstraintSchema(),
								"sqli_match_statement":                  wafv2SqliMatchStatementSchema(),
								"xss_match_statement":                   wafv2XssMatchStatementSchema(),
							},
						},
					},
				},
			},
		}
	}

	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"statement": {
					Type:     schema.TypeList,
					Required: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"byte_match_statement":                  wafv2ByteMatchStatementSchema(),
							"geo_match_statement":                   wafv2GeoMatchStatementSchema(),
							"ip_set_reference_statement":            wafv2IpSetReferenceStatementSchema(),
							"regex_pattern_set_reference_statement": wafv2RegexPatternSetReferenceStatementSchema(),
							"size_constraint_statement":             wafv2SizeConstraintSchema(),
							"sqli_match_statement":                  wafv2SqliMatchStatementSchema(),
							"xss_match_statement":                   wafv2XssMatchStatementSchema(),
						},
					},
				},
			},
		},
	}
}

func wafv2ByteMatchStatementSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"field_to_match": wafv2FieldToMatchSchema(),
				"positional_constraint": {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.StringInSlice([]string{
						wafv2.PositionalConstraintContains,
						wafv2.PositionalConstraintContainsWord,
						wafv2.PositionalConstraintEndsWith,
						wafv2.PositionalConstraintExactly,
						wafv2.PositionalConstraintStartsWith,
					}, false),
				},
				"search_string": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 200),
				},
				"text_transformation": wafv2TextTransformationSchema(),
			},
		},
	}
}

func wafv2GeoMatchStatementSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"country_codes": {
					Type:     schema.TypeList,
					Required: true,
					MinItems: 1,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"forwarded_ip_config": wafv2ForwardedIPConfig(),
			},
		},
	}
}

func wafv2IpSetReferenceStatementSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"arn": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validateArn,
				},
				"ip_set_forwarded_ip_config": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"fallback_behavior": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringInSlice(wafv2.FallbackBehavior_Values(), false),
							},
							"header_name": {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 255),
									validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]+$`), "must contain only alphanumeric and hyphen characters"),
								),
							},
							"position": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringInSlice(wafv2.ForwardedIPPosition_Values(), false),
							},
						},
					},
				},
			},
		},
	}
}

func wafv2RegexPatternSetReferenceStatementSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"arn": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validateArn,
				},
				"field_to_match":      wafv2FieldToMatchSchema(),
				"text_transformation": wafv2TextTransformationSchema(),
			},
		},
	}
}

func wafv2SizeConstraintSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"comparison_operator": {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.StringInSlice([]string{
						wafv2.ComparisonOperatorEq,
						wafv2.ComparisonOperatorGe,
						wafv2.ComparisonOperatorGt,
						wafv2.ComparisonOperatorLe,
						wafv2.ComparisonOperatorLt,
						wafv2.ComparisonOperatorNe,
					}, false),
				},
				"field_to_match": wafv2FieldToMatchSchema(),
				"size": {
					Type:         schema.TypeInt,
					Required:     true,
					ValidateFunc: validation.IntBetween(0, math.MaxInt32),
				},
				"text_transformation": wafv2TextTransformationSchema(),
			},
		},
	}
}

func wafv2SqliMatchStatementSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"field_to_match":      wafv2FieldToMatchSchema(),
				"text_transformation": wafv2TextTransformationSchema(),
			},
		},
	}
}

func wafv2XssMatchStatementSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"field_to_match":      wafv2FieldToMatchSchema(),
				"text_transformation": wafv2TextTransformationSchema(),
			},
		},
	}
}
func wafv2FieldToMatchBaseSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"all_query_arguments": wafv2EmptySchema(),
			"body":                wafv2EmptySchema(),
			"method":              wafv2EmptySchema(),
			"query_string":        wafv2EmptySchema(),
			"single_header": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 40),
								// The value is returned in lower case by the API.
								// Trying to solve it with StateFunc and/or DiffSuppressFunc resulted in hash problem of the rule field or didn't work.
								validation.StringMatch(regexp.MustCompile(`^[a-z0-9-_]+$`), "must contain only lowercase alphanumeric characters, underscores, and hyphens"),
							),
						},
					},
				},
			},
			"single_query_argument": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 30),
								// The value is returned in lower case by the API.
								// Trying to solve it with StateFunc and/or DiffSuppressFunc resulted in hash problem of the rule field or didn't work.
								validation.StringMatch(regexp.MustCompile(`^[a-z0-9-_]+$`), "must contain only lowercase alphanumeric characters, underscores, and hyphens"),
							),
						},
					},
				},
			},
			"uri_path": wafv2EmptySchema(),
		},
	}
}

func wafv2FieldToMatchSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem:     wafv2FieldToMatchBaseSchema(),
	}
}

func wafv2ForwardedIPConfig() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"fallback_behavior": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(wafv2.FallbackBehavior_Values(), false),
				},
				"header_name": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func wafv2TextTransformationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"priority": {
					Type:     schema.TypeInt,
					Required: true,
				},
				"type": {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.StringInSlice([]string{
						wafv2.TextTransformationTypeCmdLine,
						wafv2.TextTransformationTypeCompressWhiteSpace,
						wafv2.TextTransformationTypeHtmlEntityDecode,
						wafv2.TextTransformationTypeLowercase,
						wafv2.TextTransformationTypeNone,
						wafv2.TextTransformationTypeUrlDecode,
					}, false),
				},
			},
		},
	}
}

func wafv2VisibilityConfigSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cloudwatch_metrics_enabled": {
					Type:     schema.TypeBool,
					Required: true,
				},
				"metric_name": {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 128),
						validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-_]+$`), "must contain only alphanumeric hyphen and underscore characters"),
					),
				},
				"sampled_requests_enabled": {
					Type:     schema.TypeBool,
					Required: true,
				},
			},
		},
	}
}

func expandWafv2Rules(l []interface{}) ([]*wafv2.Rule, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	rules := make([]*wafv2.Rule, 0)

	for _, rule := range l {
		if rule == nil {
			continue
		}
		r, err := expandWafv2Rule(rule.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}

	return rules, nil
}

func expandWafv2Rule(m map[string]interface{}) (*wafv2.Rule, error) {
	if m == nil {
		return nil, nil
	}

	rule := &wafv2.Rule{
		Name:             aws.String(m["name"].(string)),
		Priority:         aws.Int64(int64(m["priority"].(int))),
		Action:           expandWafv2RuleAction(m["action"].([]interface{})),
		VisibilityConfig: expandWafv2VisibilityConfig(m["visibility_config"].([]interface{})),
	}
	s, err := expandWafv2RootStatement(m["statement"].([]interface{}))
	if err != nil {
		return nil, err
	}

	rule.Statement = s

	return rule, nil
}

func expandWafv2RuleAction(l []interface{}) *wafv2.RuleAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	action := &wafv2.RuleAction{}

	if v, ok := m["allow"]; ok && len(v.([]interface{})) > 0 {
		action.Allow = &wafv2.AllowAction{}
	}

	if v, ok := m["block"]; ok && len(v.([]interface{})) > 0 {
		action.Block = &wafv2.BlockAction{}
	}

	if v, ok := m["count"]; ok && len(v.([]interface{})) > 0 {
		action.Count = &wafv2.CountAction{}
	}

	return action
}

func expandWafv2VisibilityConfig(l []interface{}) *wafv2.VisibilityConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	configuration := &wafv2.VisibilityConfig{}

	if v, ok := m["cloudwatch_metrics_enabled"]; ok {
		configuration.CloudWatchMetricsEnabled = aws.Bool(v.(bool))
	}

	if v, ok := m["metric_name"]; ok && len(v.(string)) > 0 {
		configuration.MetricName = aws.String(v.(string))
	}

	if v, ok := m["sampled_requests_enabled"]; ok {
		configuration.SampledRequestsEnabled = aws.Bool(v.(bool))
	}

	return configuration
}

func expandWafv2RootStatement(l []interface{}) (*wafv2.Statement, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	m := l[0].(map[string]interface{})

	return expandWafv2Statement(m)
}

func expandWafv2Statements(l []interface{}) ([]*wafv2.Statement, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	statements := make([]*wafv2.Statement, 0)

	for _, statement := range l {
		if statement == nil {
			continue
		}
		s, err := expandWafv2Statement(statement.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		statements = append(statements, s)
	}

	return statements, nil
}

func expandWafv2Statement(m map[string]interface{}) (*wafv2.Statement, error) {
	if m == nil {
		return nil, nil
	}

	statement := &wafv2.Statement{}

	if v, ok := m["and_statement"]; ok {
		s, err := expandWafv2AndStatement(v.([]interface{}))
		if err != nil {
			return nil, err
		}
		statement.AndStatement = s
	}

	if v, ok := m["byte_match_statement"]; ok {
		s, err := expandWafv2ByteMatchStatement(v.([]interface{}))
		if err != nil {
			return nil, err
		}
		statement.ByteMatchStatement = s
	}

	if v, ok := m["ip_set_reference_statement"]; ok {
		statement.IPSetReferenceStatement = expandWafv2IpSetReferenceStatement(v.([]interface{}))
	}

	if v, ok := m["geo_match_statement"]; ok {
		statement.GeoMatchStatement = expandWafv2GeoMatchStatement(v.([]interface{}))
	}

	if v, ok := m["not_statement"]; ok {
		s, err := expandWafv2NotStatement(v.([]interface{}))
		if err != nil {
			return nil, err
		}
		statement.NotStatement = s
	}

	if v, ok := m["or_statement"]; ok {
		s, err := expandWafv2OrStatement(v.([]interface{}))
		if err != nil {
			return nil, err
		}
		statement.OrStatement = s
	}

	if v, ok := m["regex_pattern_set_reference_statement"]; ok {
		s, err := expandWafv2RegexPatternSetReferenceStatement(v.([]interface{}))
		if err != nil {
			return nil, err
		}
		statement.RegexPatternSetReferenceStatement = s
	}

	if v, ok := m["size_constraint_statement"]; ok {
		s, err := expandWafv2SizeConstraintStatement(v.([]interface{}))
		if err != nil {
			return nil, err
		}
		statement.SizeConstraintStatement = s
	}

	if v, ok := m["sqli_match_statement"]; ok {
		s, err := expandWafv2SqliMatchStatement(v.([]interface{}))
		if err != nil {
			return nil, err
		}
		statement.SqliMatchStatement = s
	}

	if v, ok := m["xss_match_statement"]; ok {
		s, err := expandWafv2XssMatchStatement(v.([]interface{}))
		if err != nil {
			return nil, err
		}
		statement.XssMatchStatement = s
	}

	return statement, nil
}

func expandWafv2AndStatement(l []interface{}) (*wafv2.AndStatement, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	m := l[0].(map[string]interface{})

	s, err := expandWafv2Statements(m["statement"].([]interface{}))
	if err != nil {
		return nil, err
	}

	return &wafv2.AndStatement{Statements: s}, nil
}

func expandWafv2ByteMatchStatement(l []interface{}) (*wafv2.ByteMatchStatement, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	m := l[0].(map[string]interface{})

	s := &wafv2.ByteMatchStatement{
		PositionalConstraint: aws.String(m["positional_constraint"].(string)),
		SearchString:         []byte(m["search_string"].(string)),
		TextTransformations:  expandWafv2TextTransformations(m["text_transformation"].(*schema.Set).List()),
	}

	fieldToMatch, err := expandWafv2FieldToMatch(m["field_to_match"].([]interface{}))
	if err != nil {
		return nil, err
	}

	s.FieldToMatch = fieldToMatch

	return s, nil
}

func expandWafv2FieldToMatch(l []interface{}) (*wafv2.FieldToMatch, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	m := l[0].(map[string]interface{})
	f := &wafv2.FieldToMatch{}

	// While the FieldToMatch struct allows more than 1 of its fields to be set,
	// the WAFv2 API does not.
	// Reference: https://github.com/terraform-providers/terraform-provider-aws/issues/14248
	nFields := 0
	for _, v := range m {
		if len(v.([]interface{})) > 0 {
			nFields++
		}
		if nFields > 1 {
			return nil, fmt.Errorf(`error expanding field_to_match: only one of "all_query_arguments", "body", method", 
							"query_string", "single_header", "single_query_argument", or "uri_path" is valid`)
		}
	}

	if v, ok := m["all_query_arguments"]; ok && len(v.([]interface{})) > 0 {
		f.AllQueryArguments = &wafv2.AllQueryArguments{}
	} else if v, ok := m["body"]; ok && len(v.([]interface{})) > 0 {
		f.Body = &wafv2.Body{}
	} else if v, ok := m["method"]; ok && len(v.([]interface{})) > 0 {
		f.Method = &wafv2.Method{}
	} else if v, ok := m["query_string"]; ok && len(v.([]interface{})) > 0 {
		f.QueryString = &wafv2.QueryString{}
	} else if v, ok := m["single_header"]; ok && len(v.([]interface{})) > 0 {
		f.SingleHeader = expandWafv2SingleHeader(m["single_header"].([]interface{}))
	} else if v, ok := m["single_query_argument"]; ok && len(v.([]interface{})) > 0 {
		f.SingleQueryArgument = expandWafv2SingleQueryArgument(m["single_query_argument"].([]interface{}))
	} else if v, ok := m["uri_path"]; ok && len(v.([]interface{})) > 0 {
		f.UriPath = &wafv2.UriPath{}
	}

	return f, nil
}

func expandWafv2ForwardedIPConfig(l []interface{}) *wafv2.ForwardedIPConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &wafv2.ForwardedIPConfig{
		FallbackBehavior: aws.String(m["fallback_behavior"].(string)),
		HeaderName:       aws.String(m["header_name"].(string)),
	}
}

func expandWafv2IPSetForwardedIPConfig(l []interface{}) *wafv2.IPSetForwardedIPConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &wafv2.IPSetForwardedIPConfig{
		FallbackBehavior: aws.String(m["fallback_behavior"].(string)),
		HeaderName:       aws.String(m["header_name"].(string)),
		Position:         aws.String(m["position"].(string)),
	}
}

func expandWafv2SingleHeader(l []interface{}) *wafv2.SingleHeader {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &wafv2.SingleHeader{
		Name: aws.String(m["name"].(string)),
	}
}

func expandWafv2SingleQueryArgument(l []interface{}) *wafv2.SingleQueryArgument {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &wafv2.SingleQueryArgument{
		Name: aws.String(m["name"].(string)),
	}
}

func expandWafv2TextTransformations(l []interface{}) []*wafv2.TextTransformation {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	rules := make([]*wafv2.TextTransformation, 0)

	for _, rule := range l {
		if rule == nil {
			continue
		}
		rules = append(rules, expandWafv2TextTransformation(rule.(map[string]interface{})))
	}

	return rules
}

func expandWafv2TextTransformation(m map[string]interface{}) *wafv2.TextTransformation {
	if m == nil {
		return nil
	}

	return &wafv2.TextTransformation{
		Priority: aws.Int64(int64(m["priority"].(int))),
		Type:     aws.String(m["type"].(string)),
	}
}

func expandWafv2IpSetReferenceStatement(l []interface{}) *wafv2.IPSetReferenceStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	statement := &wafv2.IPSetReferenceStatement{
		ARN: aws.String(m["arn"].(string)),
	}

	if v, ok := m["ip_set_forwarded_ip_config"]; ok {
		statement.IPSetForwardedIPConfig = expandWafv2IPSetForwardedIPConfig(v.([]interface{}))
	}

	return statement
}

func expandWafv2GeoMatchStatement(l []interface{}) *wafv2.GeoMatchStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	statement := &wafv2.GeoMatchStatement{
		CountryCodes: expandStringList(m["country_codes"].([]interface{})),
	}

	if v, ok := m["forwarded_ip_config"]; ok {
		statement.ForwardedIPConfig = expandWafv2ForwardedIPConfig(v.([]interface{}))
	}

	return statement
}

func expandWafv2NotStatement(l []interface{}) (*wafv2.NotStatement, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	m := l[0].(map[string]interface{})
	s := m["statement"].([]interface{})

	if len(s) == 0 || s[0] == nil {
		return nil, nil
	}

	m = s[0].(map[string]interface{})

	statement, err := expandWafv2Statement(m)
	if err != nil {
		return nil, err
	}
	return &wafv2.NotStatement{Statement: statement}, nil
}

func expandWafv2OrStatement(l []interface{}) (*wafv2.OrStatement, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	m := l[0].(map[string]interface{})

	s, err := expandWafv2Statements(m["statement"].([]interface{}))
	if err != nil {
		return nil, err
	}

	return &wafv2.OrStatement{Statements: s}, nil
}

func expandWafv2RegexPatternSetReferenceStatement(l []interface{}) (*wafv2.RegexPatternSetReferenceStatement, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	m := l[0].(map[string]interface{})

	s := &wafv2.RegexPatternSetReferenceStatement{
		ARN:                 aws.String(m["arn"].(string)),
		TextTransformations: expandWafv2TextTransformations(m["text_transformation"].(*schema.Set).List()),
	}

	fieldToMatch, err := expandWafv2FieldToMatch(m["field_to_match"].([]interface{}))
	if err != nil {
		return nil, err
	}

	s.FieldToMatch = fieldToMatch

	return s, nil
}

func expandWafv2SizeConstraintStatement(l []interface{}) (*wafv2.SizeConstraintStatement, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	m := l[0].(map[string]interface{})

	s := &wafv2.SizeConstraintStatement{
		ComparisonOperator:  aws.String(m["comparison_operator"].(string)),
		Size:                aws.Int64(int64(m["size"].(int))),
		TextTransformations: expandWafv2TextTransformations(m["text_transformation"].(*schema.Set).List()),
	}

	fieldToMatch, err := expandWafv2FieldToMatch(m["field_to_match"].([]interface{}))
	if err != nil {
		return nil, err
	}

	s.FieldToMatch = fieldToMatch

	return s, nil
}

func expandWafv2SqliMatchStatement(l []interface{}) (*wafv2.SqliMatchStatement, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	m := l[0].(map[string]interface{})

	s := &wafv2.SqliMatchStatement{
		TextTransformations: expandWafv2TextTransformations(m["text_transformation"].(*schema.Set).List()),
	}

	fieldToMatch, err := expandWafv2FieldToMatch(m["field_to_match"].([]interface{}))
	if err != nil {
		return nil, err
	}

	s.FieldToMatch = fieldToMatch

	return s, nil
}

func expandWafv2XssMatchStatement(l []interface{}) (*wafv2.XssMatchStatement, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	m := l[0].(map[string]interface{})

	s := &wafv2.XssMatchStatement{
		TextTransformations: expandWafv2TextTransformations(m["text_transformation"].(*schema.Set).List()),
	}

	fieldToMatch, err := expandWafv2FieldToMatch(m["field_to_match"].([]interface{}))
	if err != nil {
		return nil, err
	}

	s.FieldToMatch = fieldToMatch

	return s, nil
}

func flattenWafv2Rules(r []*wafv2.Rule) interface{} {
	out := make([]map[string]interface{}, len(r))
	for i, rule := range r {
		m := make(map[string]interface{})
		m["action"] = flattenWafv2RuleAction(rule.Action)
		m["name"] = aws.StringValue(rule.Name)
		m["priority"] = int(aws.Int64Value(rule.Priority))
		m["statement"] = flattenWafv2RootStatement(rule.Statement)
		m["visibility_config"] = flattenWafv2VisibilityConfig(rule.VisibilityConfig)
		out[i] = m
	}

	return out
}

func flattenWafv2RuleAction(a *wafv2.RuleAction) interface{} {
	if a == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if a.Allow != nil {
		m["allow"] = make([]map[string]interface{}, 1)
	}

	if a.Block != nil {
		m["block"] = make([]map[string]interface{}, 1)
	}

	if a.Count != nil {
		m["count"] = make([]map[string]interface{}, 1)
	}

	return []interface{}{m}
}

func flattenWafv2RootStatement(s *wafv2.Statement) interface{} {
	if s == nil {
		return []interface{}{}
	}

	return []interface{}{flattenWafv2Statement(s)}
}

func flattenWafv2Statements(s []*wafv2.Statement) interface{} {
	out := make([]interface{}, len(s))
	for i, statement := range s {
		out[i] = flattenWafv2Statement(statement)
	}

	return out
}

func flattenWafv2Statement(s *wafv2.Statement) map[string]interface{} {
	if s == nil {
		return map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if s.AndStatement != nil {
		m["and_statement"] = flattenWafv2AndStatement(s.AndStatement)
	}

	if s.ByteMatchStatement != nil {
		m["byte_match_statement"] = flattenWafv2ByteMatchStatement(s.ByteMatchStatement)
	}

	if s.IPSetReferenceStatement != nil {
		m["ip_set_reference_statement"] = flattenWafv2IpSetReferenceStatement(s.IPSetReferenceStatement)
	}

	if s.GeoMatchStatement != nil {
		m["geo_match_statement"] = flattenWafv2GeoMatchStatement(s.GeoMatchStatement)
	}

	if s.NotStatement != nil {
		m["not_statement"] = flattenWafv2NotStatement(s.NotStatement)
	}

	if s.OrStatement != nil {
		m["or_statement"] = flattenWafv2OrStatement(s.OrStatement)
	}

	if s.RegexPatternSetReferenceStatement != nil {
		m["regex_pattern_set_reference_statement"] = flattenWafv2RegexPatternSetReferenceStatement(s.RegexPatternSetReferenceStatement)
	}

	if s.SizeConstraintStatement != nil {
		m["size_constraint_statement"] = flattenWafv2SizeConstraintStatement(s.SizeConstraintStatement)
	}

	if s.SqliMatchStatement != nil {
		m["sqli_match_statement"] = flattenWafv2SqliMatchStatement(s.SqliMatchStatement)
	}

	if s.XssMatchStatement != nil {
		m["xss_match_statement"] = flattenWafv2XssMatchStatement(s.XssMatchStatement)
	}

	return m
}

func flattenWafv2AndStatement(a *wafv2.AndStatement) interface{} {
	if a == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"statement": flattenWafv2Statements(a.Statements),
	}

	return []interface{}{m}
}

func flattenWafv2ByteMatchStatement(b *wafv2.ByteMatchStatement) interface{} {
	if b == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"field_to_match":        flattenWafv2FieldToMatch(b.FieldToMatch),
		"positional_constraint": aws.StringValue(b.PositionalConstraint),
		"search_string":         string(b.SearchString),
		"text_transformation":   flattenWafv2TextTransformations(b.TextTransformations),
	}

	return []interface{}{m}
}

func flattenWafv2FieldToMatch(f *wafv2.FieldToMatch) interface{} {
	if f == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if f.AllQueryArguments != nil {
		m["all_query_arguments"] = make([]map[string]interface{}, 1)
	}

	if f.Body != nil {
		m["body"] = make([]map[string]interface{}, 1)
	}

	if f.Method != nil {
		m["method"] = make([]map[string]interface{}, 1)
	}

	if f.QueryString != nil {
		m["query_string"] = make([]map[string]interface{}, 1)
	}

	if f.SingleHeader != nil {
		m["single_header"] = flattenWafv2SingleHeader(f.SingleHeader)
	}

	if f.SingleQueryArgument != nil {
		m["single_query_argument"] = flattenWafv2SingleQueryArgument(f.SingleQueryArgument)
	}

	if f.UriPath != nil {
		m["uri_path"] = make([]map[string]interface{}, 1)
	}

	return []interface{}{m}
}

func flattenWafv2ForwardedIPConfig(f *wafv2.ForwardedIPConfig) interface{} {
	if f == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"fallback_behavior": aws.StringValue(f.FallbackBehavior),
		"header_name":       aws.StringValue(f.HeaderName),
	}

	return []interface{}{m}
}

func flattenWafv2IPSetForwardedIPConfig(i *wafv2.IPSetForwardedIPConfig) interface{} {
	if i == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"fallback_behavior": aws.StringValue(i.FallbackBehavior),
		"header_name":       aws.StringValue(i.HeaderName),
		"position":          aws.StringValue(i.Position),
	}

	return []interface{}{m}
}

func flattenWafv2SingleHeader(s *wafv2.SingleHeader) interface{} {
	if s == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"name": aws.StringValue(s.Name),
	}

	return []interface{}{m}
}

func flattenWafv2SingleQueryArgument(s *wafv2.SingleQueryArgument) interface{} {
	if s == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"name": aws.StringValue(s.Name),
	}

	return []interface{}{m}
}

func flattenWafv2TextTransformations(l []*wafv2.TextTransformation) []interface{} {
	out := make([]interface{}, len(l))
	for i, t := range l {
		m := make(map[string]interface{})
		m["priority"] = int(aws.Int64Value(t.Priority))
		m["type"] = aws.StringValue(t.Type)
		out[i] = m
	}
	return out
}

func flattenWafv2IpSetReferenceStatement(i *wafv2.IPSetReferenceStatement) interface{} {
	if i == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"arn":                        aws.StringValue(i.ARN),
		"ip_set_forwarded_ip_config": flattenWafv2IPSetForwardedIPConfig(i.IPSetForwardedIPConfig),
	}

	return []interface{}{m}
}

func flattenWafv2GeoMatchStatement(g *wafv2.GeoMatchStatement) interface{} {
	if g == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"country_codes":       flattenStringList(g.CountryCodes),
		"forwarded_ip_config": flattenWafv2ForwardedIPConfig(g.ForwardedIPConfig),
	}

	return []interface{}{m}
}

func flattenWafv2NotStatement(a *wafv2.NotStatement) interface{} {
	if a == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"statement": []interface{}{flattenWafv2Statement(a.Statement)},
	}

	return []interface{}{m}
}

func flattenWafv2OrStatement(a *wafv2.OrStatement) interface{} {
	if a == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"statement": flattenWafv2Statements(a.Statements),
	}

	return []interface{}{m}
}

func flattenWafv2RegexPatternSetReferenceStatement(r *wafv2.RegexPatternSetReferenceStatement) interface{} {
	if r == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"arn":                 aws.StringValue(r.ARN),
		"field_to_match":      flattenWafv2FieldToMatch(r.FieldToMatch),
		"text_transformation": flattenWafv2TextTransformations(r.TextTransformations),
	}

	return []interface{}{m}
}

func flattenWafv2SizeConstraintStatement(s *wafv2.SizeConstraintStatement) interface{} {
	if s == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"comparison_operator": aws.StringValue(s.ComparisonOperator),
		"field_to_match":      flattenWafv2FieldToMatch(s.FieldToMatch),
		"size":                int(aws.Int64Value(s.Size)),
		"text_transformation": flattenWafv2TextTransformations(s.TextTransformations),
	}

	return []interface{}{m}
}

func flattenWafv2SqliMatchStatement(s *wafv2.SqliMatchStatement) interface{} {
	if s == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"field_to_match":      flattenWafv2FieldToMatch(s.FieldToMatch),
		"text_transformation": flattenWafv2TextTransformations(s.TextTransformations),
	}

	return []interface{}{m}
}

func flattenWafv2XssMatchStatement(s *wafv2.XssMatchStatement) interface{} {
	if s == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"field_to_match":      flattenWafv2FieldToMatch(s.FieldToMatch),
		"text_transformation": flattenWafv2TextTransformations(s.TextTransformations),
	}

	return []interface{}{m}
}

func flattenWafv2VisibilityConfig(config *wafv2.VisibilityConfig) interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"cloudwatch_metrics_enabled": aws.BoolValue(config.CloudWatchMetricsEnabled),
		"metric_name":                aws.StringValue(config.MetricName),
		"sampled_requests_enabled":   aws.BoolValue(config.SampledRequestsEnabled),
	}

	return []interface{}{m}
}
