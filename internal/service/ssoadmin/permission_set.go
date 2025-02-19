package ssoadmin

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ssoadmin_permission_set", name="Permission Set")
// @Tags
func ResourcePermissionSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePermissionSetCreate,
		ReadWithoutTimeout:   resourcePermissionSetRead,
		UpdateWithoutTimeout: resourcePermissionSetUpdate,
		DeleteWithoutTimeout: resourcePermissionSetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 700),
					validation.StringMatch(regexp.MustCompile(`[\p{L}\p{M}\p{Z}\p{S}\p{N}\p{P}]*`), "must match [\\p{L}\\p{M}\\p{Z}\\p{S}\\p{N}\\p{P}]"),
				),
			},
			"instance_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexp.MustCompile(`[\w+=,.@-]+`), "must match [\\w+=,.@-]"),
				),
			},
			"relay_state": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 240),
					validation.StringMatch(regexp.MustCompile(`[a-zA-Z0-9&$@#\\\/%?=~\-_'"|!:,.;*+\[\]\ \(\)\{\}]+`), "must match [a-zA-Z0-9&$@#\\\\\\/%?=~\\-_'\"|!:,.;*+\\[\\]\\(\\)\\{\\}]"),
				),
			},
			"session_duration": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
				Default:      "PT1H",
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePermissionSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminConn()

	instanceARN := d.Get("instance_arn").(string)
	name := d.Get("name").(string)
	input := &ssoadmin.CreatePermissionSetInput{
		InstanceArn: aws.String(instanceARN),
		Name:        aws.String(name),
		Tags:        GetTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("relay_state"); ok {
		input.RelayState = aws.String(v.(string))
	}

	if v, ok := d.GetOk("session_duration"); ok {
		input.SessionDuration = aws.String(v.(string))
	}

	output, err := conn.CreatePermissionSetWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSO Permission Set (%s): %s", name, err)
	}

	d.SetId(fmt.Sprintf("%s,%s", aws.StringValue(output.PermissionSet.PermissionSetArn), instanceARN))

	return append(diags, resourcePermissionSetRead(ctx, d, meta)...)
}

func resourcePermissionSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminConn()

	arn, instanceARN, err := ParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing resource ID: %s", err)
	}

	output, err := conn.DescribePermissionSetWithContext(ctx, &ssoadmin.DescribePermissionSetInput{
		InstanceArn:      aws.String(instanceARN),
		PermissionSetArn: aws.String(arn),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] SSO Permission Set (%s) not found, removing from state", arn)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSO Permission Set: %s", err)
	}

	if output == nil || output.PermissionSet == nil {
		return sdkdiag.AppendErrorf(diags, "reading SSO Permission Set (%s): empty output", arn)
	}

	permissionSet := output.PermissionSet
	d.Set("arn", permissionSet.PermissionSetArn)
	d.Set("created_date", permissionSet.CreatedDate.Format(time.RFC3339))
	d.Set("description", permissionSet.Description)
	d.Set("instance_arn", instanceARN)
	d.Set("name", permissionSet.Name)
	d.Set("relay_state", permissionSet.RelayState)
	d.Set("session_duration", permissionSet.SessionDuration)

	tags, err := ListTags(ctx, conn, arn, instanceARN)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for SSO Permission Set (%s): %s", arn, err)
	}

	SetTagsOut(ctx, Tags(tags))

	return diags
}

func resourcePermissionSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminConn()

	arn, instanceARN, err := ParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing resource ID: %s", err)
	}

	if d.HasChanges("description", "relay_state", "session_duration") {
		input := &ssoadmin.UpdatePermissionSetInput{
			InstanceArn:      aws.String(instanceARN),
			PermissionSetArn: aws.String(arn),
		}

		// The AWS SSO API requires we send the RelayState value regardless if it's unchanged
		// else the existing Permission Set's RelayState value will be cleared;
		// for consistency, we'll check for the "presence of" instead of "if changed" for all input fields
		// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/17411

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("relay_state"); ok {
			input.RelayState = aws.String(v.(string))
		}

		if v, ok := d.GetOk("session_duration"); ok {
			input.SessionDuration = aws.String(v.(string))
		}

		_, err := conn.UpdatePermissionSetWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SSO Permission Set (%s): %s", arn, err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, arn, instanceARN, o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags: %s", err)
		}
	}

	// Re-provision ALL accounts after making the above changes
	if err := provisionPermissionSet(ctx, conn, arn, instanceARN); err != nil {
		return sdkdiag.AppendErrorf(diags, "provisioning SSO Permission Set (%s): %s", arn, err)
	}

	return append(diags, resourcePermissionSetRead(ctx, d, meta)...)
}

func resourcePermissionSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminConn()

	arn, instanceARN, err := ParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing resource ID: %s", err)
	}

	log.Printf("[INFO] Deleting SSO Permission Set: %s", d.Id())
	_, err = conn.DeletePermissionSetWithContext(ctx, &ssoadmin.DeletePermissionSetInput{
		InstanceArn:      aws.String(instanceARN),
		PermissionSetArn: aws.String(arn),
	})

	if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSO Permission Set (%s): %s", arn, err)
	}

	return diags
}

func ParseResourceID(id string) (string, string, error) {
	idParts := strings.Split(id, ",")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%q), expected PERMISSION_SET_ARN,INSTANCE_ARN", id)
	}
	return idParts[0], idParts[1], nil
}

func provisionPermissionSet(ctx context.Context, conn *ssoadmin.SSOAdmin, arn, instanceArn string) error {
	input := &ssoadmin.ProvisionPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(arn),
		TargetType:       aws.String(ssoadmin.ProvisionTargetTypeAllProvisionedAccounts),
	}

	var output *ssoadmin.ProvisionPermissionSetOutput
	err := retry.RetryContext(ctx, permissionSetProvisionTimeout, func() *retry.RetryError {
		var err error
		output, err = conn.ProvisionPermissionSetWithContext(ctx, input)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeConflictException) {
				return retry.RetryableError(err)
			}
			if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeThrottlingException) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.ProvisionPermissionSetWithContext(ctx, input)
	}

	if err != nil {
		return fmt.Errorf("provisioning SSO Permission Set (%s): %w", arn, err)
	}

	if output == nil || output.PermissionSetProvisioningStatus == nil {
		return fmt.Errorf("provisioning SSO Permission Set (%s): empty output", arn)
	}

	_, err = waitPermissionSetProvisioned(ctx, conn, instanceArn, aws.StringValue(output.PermissionSetProvisioningStatus.RequestId))
	if err != nil {
		return fmt.Errorf("waiting for SSO Permission Set (%s) to provision: %w", arn, err)
	}

	return nil
}
