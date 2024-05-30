// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdb

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_docdb_cluster", name="Cluster")
// @Tags(identifierAttribute="arn")
func ResourceCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterCreate,
		ReadWithoutTimeout:   resourceClusterRead,
		UpdateWithoutTimeout: resourceClusterUpdate,
		DeleteWithoutTimeout: resourceClusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				// Neither skip_final_snapshot nor final_snapshot_identifier can be fetched
				// from any API call, so we need to default skip_final_snapshot to true so
				// that final_snapshot_identifier is not required
				d.Set("skip_final_snapshot", true)
				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"allow_major_version_upgrade": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrApplyImmediately: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAvailabilityZones: {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"backup_retention_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntAtMost(35),
			},
			names.AttrClusterIdentifier: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"cluster_identifier_prefix"},
				ValidateFunc:  validIdentifier,
			},
			"cluster_identifier_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrClusterIdentifier},
				ValidateFunc:  validIdentifierPrefix,
			},
			"cluster_members": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Computed: true,
			},
			"cluster_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_cluster_parameter_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"db_subnet_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrDeletionProtection: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"enabled_cloudwatch_logs_exports": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"audit",
						"profiler",
					}, false),
				},
			},
			names.AttrEndpoint: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEngine: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      engineDocDB,
				ValidateFunc: validation.StringInSlice(engine_Values(), false),
			},
			names.AttrEngineVersion: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"final_snapshot_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
					value := v.(string)
					if !regexache.MustCompile(`^[0-9A-Za-z-]+$`).MatchString(value) {
						es = append(es, fmt.Errorf(
							"only alphanumeric characters and hyphens allowed in %q", k))
					}
					if regexache.MustCompile(`--`).MatchString(value) {
						es = append(es, fmt.Errorf("%q cannot contain two consecutive hyphens", k))
					}
					if regexache.MustCompile(`-$`).MatchString(value) {
						es = append(es, fmt.Errorf("%q cannot end in a hyphen", k))
					}
					return
				},
			},
			"global_cluster_identifier": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validGlobalCusterIdentifier,
			},
			names.AttrHostedZoneID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"master_password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"master_username": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrPort: {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      27017,
				ValidateFunc: validation.IntBetween(1150, 65535),
			},
			"preferred_backup_window": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidOnceADayWindowFormat,
			},
			names.AttrPreferredMaintenanceWindow: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				StateFunc: func(val interface{}) string {
					if val == nil {
						return ""
					}
					return strings.ToLower(val.(string))
				},
				ValidateFunc: verify.ValidOnceAWeekWindowFormat,
			},
			"reader_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"skip_final_snapshot": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"snapshot_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// allow snapshot_idenfitier to be removed without forcing re-creation
					return new == ""
				},
			},
			names.AttrStorageEncrypted: {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			names.AttrStorageType: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(storageType_Values(), false),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// When you create a DocumentDB DB cluster with the storage type set to "iopt1",
					// the storage type is returned in the response.
					// The storage type isn't returned when you set it to "standard".
					if old == "" && new == storageTypeStandard {
						return true
					}
					return old == new
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBConn(ctx)

	identifier := create.NewNameGenerator(
		create.WithConfiguredName(d.Get(names.AttrClusterIdentifier).(string)),
		create.WithConfiguredPrefix(d.Get("cluster_identifier_prefix").(string)),
		create.WithDefaultPrefix("tf-"),
	).Generate()

	// Some API calls (e.g. RestoreDBClusterFromSnapshot do not support all
	// parameters to correctly apply all settings in one pass. For missing
	// parameters or unsupported configurations, we may need to call
	// ModifyDBInstance afterwadocdb to prevent Terraform operators from API
	// errors or needing to double apply.
	var requiresModifyDbCluster bool
	inputM := &docdb.ModifyDBClusterInput{
		ApplyImmediately: aws.Bool(true),
	}

	if _, ok := d.GetOk("snapshot_identifier"); ok {
		input := &docdb.RestoreDBClusterFromSnapshotInput{
			DBClusterIdentifier: aws.String(identifier),
			DeletionProtection:  aws.Bool(d.Get(names.AttrDeletionProtection).(bool)),
			Engine:              aws.String(d.Get(names.AttrEngine).(string)),
			SnapshotIdentifier:  aws.String(d.Get("snapshot_identifier").(string)),
			Tags:                getTagsIn(ctx),
		}

		if v := d.Get(names.AttrAvailabilityZones).(*schema.Set); v.Len() > 0 {
			input.AvailabilityZones = flex.ExpandStringSet(v)
		}

		if v, ok := d.GetOk("backup_retention_period"); ok {
			inputM.BackupRetentionPeriod = aws.Int64(int64(v.(int)))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("db_cluster_parameter_group_name"); ok {
			inputM.DBClusterParameterGroupName = aws.String(v.(string))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk("db_subnet_group_name"); ok {
			input.DBSubnetGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && len(v.([]interface{})) > 0 {
			input.EnableCloudwatchLogsExports = flex.ExpandStringList(v.([]interface{}))
		}

		if v, ok := d.GetOk(names.AttrEngineVersion); ok {
			input.EngineVersion = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
			input.KmsKeyId = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrPort); ok {
			input.Port = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("preferred_backup_window"); ok {
			inputM.PreferredBackupWindow = aws.String(v.(string))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk(names.AttrPreferredMaintenanceWindow); ok {
			inputM.PreferredMaintenanceWindow = aws.String(v.(string))
			requiresModifyDbCluster = true
		}

		if v, ok := d.GetOk(names.AttrStorageType); ok {
			input.StorageType = aws.String(v.(string))
		}

		if v := d.Get(names.AttrVPCSecurityGroupIDs).(*schema.Set); v.Len() > 0 {
			input.VpcSecurityGroupIds = flex.ExpandStringSet(v)
		}

		_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
			return conn.RestoreDBClusterFromSnapshotWithContext(ctx, input)
		}, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions")

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating DocumentDB Cluster (restore from snapshot) (%s): %s", identifier, err)
		}
	} else {
		// Secondary DocDB clusters part of a global cluster will not supply the master_password
		if _, ok := d.GetOk("global_cluster_identifier"); !ok {
			if _, ok := d.GetOk("master_password"); !ok {
				return sdkdiag.AppendErrorf(diags, `provider.aws: aws_docdb_cluster: %s: "master_password": required field is not set`, identifier)
			}
		}

		// Secondary DocDB clusters part of a global cluster will not supply the master_username
		if _, ok := d.GetOk("global_cluster_identifier"); !ok {
			if _, ok := d.GetOk("master_username"); !ok {
				return sdkdiag.AppendErrorf(diags, `provider.aws: aws_docdb_cluster: %s: "master_username": required field is not set`, identifier)
			}
		}

		input := &docdb.CreateDBClusterInput{
			DBClusterIdentifier: aws.String(identifier),
			DeletionProtection:  aws.Bool(d.Get(names.AttrDeletionProtection).(bool)),
			Engine:              aws.String(d.Get(names.AttrEngine).(string)),
			MasterUsername:      aws.String(d.Get("master_username").(string)),
			MasterUserPassword:  aws.String(d.Get("master_password").(string)),
			Tags:                getTagsIn(ctx),
		}

		if v := d.Get(names.AttrAvailabilityZones).(*schema.Set); v.Len() > 0 {
			input.AvailabilityZones = flex.ExpandStringSet(v)
		}

		if v, ok := d.GetOk("backup_retention_period"); ok {
			input.BackupRetentionPeriod = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("db_cluster_parameter_group_name"); ok {
			input.DBClusterParameterGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("db_subnet_group_name"); ok {
			input.DBSubnetGroupName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("enabled_cloudwatch_logs_exports"); ok && len(v.([]interface{})) > 0 {
			input.EnableCloudwatchLogsExports = flex.ExpandStringList(v.([]interface{}))
		}

		if v, ok := d.GetOk(names.AttrEngineVersion); ok {
			input.EngineVersion = aws.String(v.(string))
		}

		if v, ok := d.GetOk("global_cluster_identifier"); ok {
			input.GlobalClusterIdentifier = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
			input.KmsKeyId = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrPort); ok {
			input.Port = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("preferred_backup_window"); ok {
			input.PreferredBackupWindow = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrPreferredMaintenanceWindow); ok {
			input.PreferredMaintenanceWindow = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrStorageEncrypted); ok {
			input.StorageEncrypted = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk(names.AttrStorageType); ok {
			input.StorageType = aws.String(v.(string))
		}

		if v := d.Get(names.AttrVPCSecurityGroupIDs).(*schema.Set); v.Len() > 0 {
			input.VpcSecurityGroupIds = flex.ExpandStringSet(v)
		}

		_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
			return conn.CreateDBClusterWithContext(ctx, input)
		}, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions")

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating DocumentDB Cluster (%s): %s", identifier, err)
		}
	}

	d.SetId(identifier)

	if _, err := waitDBClusterAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DocumentDB Cluster (%s) create: %s", d.Id(), err)
	}

	if requiresModifyDbCluster {
		inputM.DBClusterIdentifier = aws.String(d.Id())

		_, err := conn.ModifyDBClusterWithContext(ctx, inputM)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying DocumentDB Cluster (%s): %s", d.Id(), err)
		}

		if _, err := waitDBClusterAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for DocumentDB Cluster (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBConn(ctx)

	dbc, err := FindDBClusterByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DocumentDB Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DocumentDB Cluster (%s): %s", d.Id(), err)
	}

	// Ignore the following API error for regions/partitions that do not support DocDB Global Clusters:
	// InvalidParameterValue: Access Denied to API Version: APIGlobalDatabases
	if globalCluster, err := findGlobalClusterByClusterARN(ctx, conn, aws.StringValue(dbc.DBClusterArn)); tfresource.NotFound(err) || tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "Access Denied to API Version: APIGlobalDatabases") {
		d.Set("global_cluster_identifier", "")
	} else if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DocumentDB Global Cluster information for DocumentDB Cluster (%s): %s", d.Id(), err)
	} else {
		d.Set("global_cluster_identifier", globalCluster.GlobalClusterIdentifier)
	}

	d.Set(names.AttrARN, dbc.DBClusterArn)
	d.Set(names.AttrAvailabilityZones, aws.StringValueSlice(dbc.AvailabilityZones))
	d.Set("backup_retention_period", dbc.BackupRetentionPeriod)
	d.Set(names.AttrClusterIdentifier, dbc.DBClusterIdentifier)
	d.Set("cluster_identifier_prefix", create.NamePrefixFromName(aws.StringValue(dbc.DBClusterIdentifier)))
	var clusterMembers []string
	for _, v := range dbc.DBClusterMembers {
		clusterMembers = append(clusterMembers, aws.StringValue(v.DBInstanceIdentifier))
	}
	d.Set("cluster_members", clusterMembers)
	d.Set("cluster_resource_id", dbc.DbClusterResourceId)
	d.Set("db_cluster_parameter_group_name", dbc.DBClusterParameterGroup)
	d.Set("db_subnet_group_name", dbc.DBSubnetGroup)
	d.Set(names.AttrDeletionProtection, dbc.DeletionProtection)
	d.Set("enabled_cloudwatch_logs_exports", aws.StringValueSlice(dbc.EnabledCloudwatchLogsExports))
	d.Set(names.AttrEndpoint, dbc.Endpoint)
	d.Set(names.AttrEngineVersion, dbc.EngineVersion)
	d.Set(names.AttrEngine, dbc.Engine)
	d.Set(names.AttrHostedZoneID, dbc.HostedZoneId)
	d.Set(names.AttrKMSKeyID, dbc.KmsKeyId)
	d.Set("master_username", dbc.MasterUsername)
	d.Set(names.AttrPort, dbc.Port)
	d.Set("preferred_backup_window", dbc.PreferredBackupWindow)
	d.Set(names.AttrPreferredMaintenanceWindow, dbc.PreferredMaintenanceWindow)
	d.Set("reader_endpoint", dbc.ReaderEndpoint)
	d.Set(names.AttrStorageEncrypted, dbc.StorageEncrypted)
	d.Set(names.AttrStorageType, dbc.StorageType)
	var securityGroupIDs []string
	for _, v := range dbc.VpcSecurityGroups {
		securityGroupIDs = append(securityGroupIDs, aws.StringValue(v.VpcSecurityGroupId))
	}
	d.Set(names.AttrVPCSecurityGroupIDs, securityGroupIDs)

	return diags
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll, "global_cluster_identifier", "skip_final_snapshot") {
		input := &docdb.ModifyDBClusterInput{
			ApplyImmediately:    aws.Bool(d.Get(names.AttrApplyImmediately).(bool)),
			DBClusterIdentifier: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("allow_major_version_upgrade"); ok {
			input.AllowMajorVersionUpgrade = aws.Bool(v.(bool))
		}

		if d.HasChange("backup_retention_period") {
			input.BackupRetentionPeriod = aws.Int64(int64(d.Get("backup_retention_period").(int)))
		}

		if d.HasChange("db_cluster_parameter_group_name") {
			input.DBClusterParameterGroupName = aws.String(d.Get("db_cluster_parameter_group_name").(string))
		}

		if d.HasChange(names.AttrDeletionProtection) {
			input.DeletionProtection = aws.Bool(d.Get(names.AttrDeletionProtection).(bool))
		}

		if d.HasChange("enabled_cloudwatch_logs_exports") {
			input.CloudwatchLogsExportConfiguration = expandCloudwatchLogsExportConfiguration(d)
		}

		if d.HasChange(names.AttrEngineVersion) {
			input.EngineVersion = aws.String(d.Get(names.AttrEngineVersion).(string))
		}

		if d.HasChange("master_password") {
			input.MasterUserPassword = aws.String(d.Get("master_password").(string))
		}

		if d.HasChange("preferred_backup_window") {
			input.PreferredBackupWindow = aws.String(d.Get("preferred_backup_window").(string))
		}

		if d.HasChange(names.AttrStorageType) {
			input.StorageType = aws.String(d.Get(names.AttrStorageType).(string))
		}

		if d.HasChange(names.AttrPreferredMaintenanceWindow) {
			input.PreferredMaintenanceWindow = aws.String(d.Get(names.AttrPreferredMaintenanceWindow).(string))
		}

		if d.HasChange(names.AttrVPCSecurityGroupIDs) {
			if v := d.Get(names.AttrVPCSecurityGroupIDs).(*schema.Set); v.Len() > 0 {
				input.VpcSecurityGroupIds = flex.ExpandStringSet(v)
			} else {
				input.VpcSecurityGroupIds = aws.StringSlice([]string{})
			}
		}

		_, err := tfresource.RetryWhen(ctx, 5*time.Minute,
			func() (interface{}, error) {
				return conn.ModifyDBClusterWithContext(ctx, input)
			},
			func(err error) (bool, error) {
				if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions") {
					return true, err
				}

				if tfawserr.ErrMessageContains(err, docdb.ErrCodeInvalidDBClusterStateFault, "is not currently in the available state") {
					return true, err
				}

				if tfawserr.ErrMessageContains(err, docdb.ErrCodeInvalidDBClusterStateFault, "cluster is a part of a global cluster") {
					return true, err
				}

				return false, err
			},
		)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying DocumentDB Cluster (%s): %s", d.Id(), err)
		}

		if _, err := waitDBClusterAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for DocumentDB Cluster (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("global_cluster_identifier") {
		oRaw, nRaw := d.GetChange("global_cluster_identifier")
		o, n := oRaw.(string), nRaw.(string)

		if o == "" {
			return sdkdiag.AppendErrorf(diags, "existing DocumentDB Clusters cannot be added to an existing DocumentDB Global Cluster")
		}

		if n != "" {
			return sdkdiag.AppendErrorf(diags, "existing DocumentDB Clusters cannot be migrated between existing DocumentDB Global Clusters")
		}

		if err := removeClusterFromGlobalCluster(ctx, conn, d.Get(names.AttrARN).(string), o, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBConn(ctx)

	skipFinalSnapshot := d.Get("skip_final_snapshot").(bool)
	input := &docdb.DeleteDBClusterInput{
		DBClusterIdentifier: aws.String(d.Id()),
		SkipFinalSnapshot:   aws.Bool(skipFinalSnapshot),
	}

	if !skipFinalSnapshot {
		if v, ok := d.GetOk("final_snapshot_identifier"); ok {
			input.FinalDBSnapshotIdentifier = aws.String(v.(string))
		} else {
			return sdkdiag.AppendErrorf(diags, "DocumentDB Cluster FinalSnapshotIdentifier is required when a final snapshot is required")
		}
	}

	if v, ok := d.GetOk("global_cluster_identifier"); ok {
		if err := removeClusterFromGlobalCluster(ctx, conn, d.Get(names.AttrARN).(string), v.(string), d.Timeout(schema.TimeoutDelete)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[DEBUG] Deleting DocumentDB Cluster: %s", d.Id())
	_, err := tfresource.RetryWhen(ctx, d.Timeout(schema.TimeoutDelete),
		func() (interface{}, error) {
			return conn.DeleteDBClusterWithContext(ctx, input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, docdb.ErrCodeInvalidDBClusterStateFault, "is not currently in the available state") {
				return true, err
			}

			if tfawserr.ErrMessageContains(err, docdb.ErrCodeInvalidDBClusterStateFault, "cluster is a part of a global cluster") {
				return true, err
			}

			return false, err
		},
	)

	if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBClusterNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DocumentDB Cluster (%s): %s", d.Id(), err)
	}

	if _, err := waitDBClusterDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DocumentDB Cluster (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandCloudwatchLogsExportConfiguration(d *schema.ResourceData) *docdb.CloudwatchLogsExportConfiguration { // nosemgrep:ci.caps0-in-func-name
	oraw, nraw := d.GetChange("enabled_cloudwatch_logs_exports")
	o := oraw.([]interface{})
	n := nraw.([]interface{})

	create, disable := diffCloudWatchLogsExportConfiguration(o, n)

	return &docdb.CloudwatchLogsExportConfiguration{
		EnableLogTypes:  flex.ExpandStringList(create),
		DisableLogTypes: flex.ExpandStringList(disable),
	}
}

func diffCloudWatchLogsExportConfiguration(old, new []interface{}) ([]interface{}, []interface{}) {
	add := make([]interface{}, 0)
	disable := make([]interface{}, 0)

	for _, n := range new {
		if idx := tfslices.IndexOf(old, n.(string)); idx == -1 {
			add = append(add, n)
		}
	}

	for _, o := range old {
		if idx := tfslices.IndexOf(new, o.(string)); idx == -1 {
			disable = append(disable, o)
		}
	}

	return add, disable
}

func removeClusterFromGlobalCluster(ctx context.Context, conn *docdb.DocDB, clusterARN, globalClusterID string, timeout time.Duration) error {
	input := &docdb.RemoveFromGlobalClusterInput{
		DbClusterIdentifier:     aws.String(clusterARN),
		GlobalClusterIdentifier: aws.String(globalClusterID),
	}

	_, err := conn.RemoveFromGlobalClusterWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBClusterNotFoundFault, docdb.ErrCodeGlobalClusterNotFoundFault) || tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "is not found in global cluster") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("removing DocumentDB Cluster (%s) from DocumentDB Global Cluster (%s): %w", clusterARN, globalClusterID, err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, timeout, func() (interface{}, error) {
		return findGlobalClusterByClusterARN(ctx, conn, clusterARN)
	})

	if err != nil {
		return fmt.Errorf("waiting for DocumentDB Cluster (%s) removal from DocumentDB Global Cluster (%s): %w", clusterARN, globalClusterID, err)
	}

	return nil
}

func FindDBClusterByID(ctx context.Context, conn *docdb.DocDB, id string) (*docdb.DBCluster, error) {
	input := &docdb.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(id),
	}
	output, err := findDBCluster(ctx, conn, input, tfslices.PredicateTrue[*docdb.DBCluster]())

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.DBClusterIdentifier) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findClusterByARN(ctx context.Context, conn *docdb.DocDB, arn string) (*docdb.DBCluster, error) {
	input := &docdb.DescribeDBClustersInput{}

	return findDBCluster(ctx, conn, input, func(v *docdb.DBCluster) bool {
		return aws.StringValue(v.DBClusterArn) == arn
	})
}

func findDBCluster(ctx context.Context, conn *docdb.DocDB, input *docdb.DescribeDBClustersInput, filter tfslices.Predicate[*docdb.DBCluster]) (*docdb.DBCluster, error) {
	output, err := findDBClusters(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findDBClusters(ctx context.Context, conn *docdb.DocDB, input *docdb.DescribeDBClustersInput, filter tfslices.Predicate[*docdb.DBCluster]) ([]*docdb.DBCluster, error) {
	var output []*docdb.DBCluster

	err := conn.DescribeDBClustersPagesWithContext(ctx, input, func(page *docdb.DescribeDBClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBClusters {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBClusterNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func statusDBCluster(ctx context.Context, conn *docdb.DocDB, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDBClusterByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitDBClusterAvailable(ctx context.Context, conn *docdb.DocDB, id string, timeout time.Duration) (*docdb.DBCluster, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			clusterStatusCreating,
			clusterStatusBackingUp,
			clusterStatusModifying,
			clusterStatusPreparingDataMigration,
			clusterStatusMigrating,
			clusterStatusResettingMasterCredentials,
			clusterStatusUpgrading,
		},
		Target:     []string{clusterStatusAvailable},
		Refresh:    statusDBCluster(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*docdb.DBCluster); ok {
		return output, err
	}

	return nil, err
}

func waitDBClusterDeleted(ctx context.Context, conn *docdb.DocDB, id string, timeout time.Duration) (*docdb.DBCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			clusterStatusAvailable,
			clusterStatusDeleting,
			clusterStatusBackingUp,
			clusterStatusModifying,
		},
		Target:     []string{},
		Refresh:    statusDBCluster(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*docdb.DBCluster); ok {
		return output, err
	}

	return nil, err
}
