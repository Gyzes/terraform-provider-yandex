package yandex

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	sdkoperation "github.com/yandex-cloud/go-sdk/operation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"google.golang.org/genproto/protobuf/field_mask"
)

const (
	yandexMDBOpenSearchClusterCreateTimeout    = 30 * time.Minute
	yandexMDBOpenSearchClusterDeleteTimeout    = 15 * time.Minute
	yandexMDBOpenSearchClusterUpdateTimeout    = 60 * time.Minute
	yandexMDBOpenSearchOperationsRetryCount    = 5
	yandexMDBOpenSearchOperationsRetryInterval = 2 * time.Minute
)

func resourceYandexMDBOpenSearchCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexMDBOpenSearchClusterCreate,
		ReadContext:   resourceYandexMDBOpenSearchClusterRead,
		UpdateContext: resourceYandexMDBOpenSearchClusterUpdate,
		DeleteContext: resourceYandexMDBOpenSearchClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBOpenSearchClusterCreateTimeout),
			Update: schema.DefaultTimeout(yandexMDBOpenSearchClusterUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexMDBOpenSearchClusterDeleteTimeout),
		},

		CustomizeDiff: opensearchNodeGroupsDiffCustomize,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			// ID of the folder that the OpenSearch cluster belongs to.
			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true, // TODO impl move cluster
			},

			// Creation timestamp.
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			// Name of the OpenSearch cluster. The name must be unique within the folder.
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			// Description of the OpenSearch cluster.
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			// Custom labels for the OpenSearch cluster as `key:value` pairs.
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			// Deployment environment of the OpenSearch cluster.
			"environment": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			// Configuration of the OpenSearch cluster.
			"config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"version": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},

						"admin_password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},

						"opensearch": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"node_groups": {
										Type:     schema.TypeSet,
										MinItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Required: true,
												},

												"resources": {
													Type:     schema.TypeSet,
													Required: true,
													MaxItems: 1,
													MinItems: 1,
													//Computed: true,
													//Optional: true,
													Elem: openSearchResourcesSchema(),
												},

												"hosts_count": {
													Type:     schema.TypeInt,
													Required: true,
												},

												"zone_ids": {
													Type:     schema.TypeSet,
													Elem:     &schema.Schema{Type: schema.TypeString},
													Set:      schema.HashString,
													Required: true,
												},

												"subnet_ids": {
													Type:     schema.TypeSet,
													Elem:     &schema.Schema{Type: schema.TypeString},
													Set:      schema.HashString,
													Optional: true,
													Computed: true,
												},

												"assign_public_ip": {
													Type:     schema.TypeBool,
													Computed: true,
													Optional: true,
												},

												"roles": {
													Type: schema.TypeSet,
													Elem: &schema.Schema{Type: schema.TypeString},
													//Set:      schema.HashString,
													Set:      openSearchRoleHash,
													Optional: true,
													Computed: true,
												},
											},
										},
										Set: openSearchNodeGroupDeepHash,
										//Set:      openSearchNodeGroupNameHash,
										Required: true,
										//Optional: true,
										//Computed: true,
									},

									"plugins": {
										Type:     schema.TypeSet,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Set:      schema.HashString,
										Optional: true,
										Computed: true,
									},
								},
							},
						},

						"dashboards": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"node_groups": {
										Type:     schema.TypeSet,
										MinItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Required: true,
												},

												"resources": {
													Type:     schema.TypeSet,
													Required: true,
													MaxItems: 1,
													MinItems: 1,
													Elem:     openSearchResourcesSchema(),
												},

												"hosts_count": {
													Type:     schema.TypeInt,
													Required: true,
												},

												"zone_ids": {
													Type:     schema.TypeSet,
													Elem:     &schema.Schema{Type: schema.TypeString},
													Set:      schema.HashString,
													Required: true,
												},

												"subnet_ids": {
													Type:     schema.TypeSet,
													Elem:     &schema.Schema{Type: schema.TypeString},
													Set:      schema.HashString,
													Optional: true,
													Computed: true,
												},

												"assign_public_ip": {
													Type:     schema.TypeBool,
													Computed: true,
													Optional: true,
												},
											},
										},
										//Set: dashboardsNodeGroupDeepHash,

										//Set:      dashboardsNodeGroupNameHash,
										//Optional: true,
										//Computed: true,
										Required: true,
									},
								},
							},
						},
					},
				},
			},

			// ID of the network that the cluster belongs to.
			"network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			// Aggregated cluster health.
			"health": {
				Type:     schema.TypeString,
				Computed: true,
			},

			// Current state of the cluster.
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			// Current nodes in the cluster
			"hosts": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      opensearchHostFQDNHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fqdn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"zone": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"roles": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"assign_public_ip": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},

			// User security groups
			"security_group_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
			},

			"service_account_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"maintenance_window": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{"ANYTIME", "WEEKLY"}, false),
							Required:     true,
						},
						"day": {
							Type:         schema.TypeString,
							ValidateFunc: validateParsableValue(parseOpenSearchWeekDay),
							Optional:     true,
						},
						"hour": {
							Type:         schema.TypeInt,
							ValidateFunc: validation.IntBetween(1, 24),
							Optional:     true,
						},
					},
				},
			},
		},
	}
}

func openSearchResourcesSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"resource_preset_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"disk_size": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"disk_type_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceYandexMDBOpenSearchClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if err := resourceYandexMDBOpenSearchClusterReadEx(ctx, d, meta, "ResourceRead"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceYandexMDBOpenSearchClusterReadEx(ctx context.Context, d *schema.ResourceData, meta interface{}, from string) error {
	config := meta.(*Config)

	tflog.Debug(ctx, "Reading OpenSearch Cluster", map[string]interface{}{"id": d.Id()})

	cluster, err := config.sdk.MDB().OpenSearch().Cluster().Get(ctx, &opensearch.GetClusterRequest{
		ClusterId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Id()))
	}

	d.Set("folder_id", cluster.GetFolderId())
	d.Set("created_at", getTimestamp(cluster.CreatedAt))

	d.Set("name", cluster.GetName())
	d.Set("description", cluster.GetDescription())
	if err := d.Set("labels", cluster.GetLabels()); err != nil {
		return err
	}

	d.Set("environment", cluster.GetEnvironment().String())

	password := ""
	if v, ok := d.GetOk("config.0.admin_password"); ok {
		password = v.(string)
	}
	clusterConfig := flattenOpenSearchClusterConfig(cluster.Config, password)
	if err := d.Set("config", clusterConfig); err != nil {
		return err
	}

	d.Set("network_id", cluster.GetNetworkId())

	d.Set("health", cluster.GetHealth().String())

	d.Set("status", cluster.GetStatus().String())

	if cluster.SecurityGroupIds == nil {
		cluster.SecurityGroupIds = []string{}
	}
	if err := d.Set("security_group_ids", cluster.SecurityGroupIds); err != nil {
		return err
	}

	actualHosts, err := listOpensearchHosts(ctx, config, d.Id())
	if err != nil {
		return err
	}

	result := flattenOpensearchHosts(actualHosts)

	if err := d.Set("hosts", result); err != nil {
		return err
	}

	d.Set("service_account_id", cluster.GetServiceAccountId())

	d.Set("deletion_protection", cluster.GetDeletionProtection())

	mw := flattenOpenSearchMaintenanceWindow(cluster.MaintenanceWindow)
	if err := d.Set("maintenance_window", mw); err != nil {
		return err
	}

	tflog.Debug(ctx, "Finished reading OpenSearch Cluster", map[string]interface{}{"id": d.Id()})

	return nil
}

func shouldRetry(op *sdkoperation.Operation, err error) bool {
	if err != nil {
		status, ok := status.FromError(err)
		if ok && status.Code() == codes.Internal {
			return true
		}
	} else if op.Failed() {
		return true
	}
	return false
}

func waitOperationWithRetry(ctx context.Context, config *Config, caller string, action func() (*operation.Operation, error)) error {
	retryCount := yandexMDBOpenSearchOperationsRetryCount
	var err error
	for ; retryCount > 0; retryCount-- {
		var op *sdkoperation.Operation
		op, err := retryConflictingOperation(ctx, config, action)
		if err != nil {
			return fmt.Errorf("error while requesting API for %s: %w", caller, err)
		}

		err = op.Wait(ctx)
		if shouldRetry(op, err) {
			time.Sleep(yandexMDBOpenSearchOperationsRetryInterval)
			continue
		}
		_, err = op.Response()
		if shouldRetry(op, err) {
			time.Sleep(yandexMDBOpenSearchOperationsRetryInterval)
			continue
		}
		break
	}
	return err
}

func resourceYandexMDBOpenSearchClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	tflog.Debug(ctx, "Creating OpenSearch Cluster")

	req, err := prepareCreateOpenSearchRequest(d, config)

	if err != nil {
		return diag.FromErr(err)
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().OpenSearch().Cluster().Create(ctx, req))
	if err != nil {
		return diag.Errorf("Error while requesting API to create OpenSearch Cluster: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return diag.Errorf("Error while get OpenSearch Cluster create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*opensearch.CreateClusterMetadata)
	if !ok {
		return diag.Errorf("Could not get OpenSearch Cluster ID from create operation metadata")
	}

	d.SetId(md.ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		return diag.Errorf("Error while waiting for operation to create OpenSearch Cluster: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return diag.FromErr(fmt.Errorf("OpenSearch Cluster creation failed: %w", err))
	}

	tflog.Debug(ctx, "Finishing creating OpenSearch Cluster", map[string]interface{}{"id": d.Id()})

	return resourceYandexMDBOpenSearchClusterRead(ctx, d, meta)
}

func prepareCreateOpenSearchRequest(d *schema.ResourceData, meta *Config) (*opensearch.CreateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error while expanding labels on OpenSearch Cluster create: %w", err)
	}

	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error getting folder ID while creating OpenSearch Cluster: %w", err)
	}

	e := d.Get("environment").(string)
	env, err := parseOpenSearchEnv(e)
	if err != nil {
		return nil, fmt.Errorf("Error resolving environment while creating OpenSearch Cluster: %w", err)
	}

	securityGroupIds := expandSecurityGroupIds(d.Get("security_group_ids"))

	config := expandOpenSearchConfigCreateSpec(d.Get("config"))

	networkID, err := expandAndValidateNetworkId(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding network id on OpenSearch Cluster create: %w", err)
	}

	mntWindow, err := expandOpenSearchMaintenanceWindow(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding maintenance window on OpenSearch Cluster create: %w", err)
	}

	req := &opensearch.CreateClusterRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,

		FolderId:    folderID,
		Environment: env,
		NetworkId:   networkID,

		ConfigSpec: config,

		SecurityGroupIds:   securityGroupIds,
		ServiceAccountId:   d.Get("service_account_id").(string),
		DeletionProtection: d.Get("deletion_protection").(bool),
		MaintenanceWindow:  mntWindow,
	}

	return req, nil
}

func resourceYandexMDBOpenSearchClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	tflog.Debug(ctx, "Deleting OpenSearch Cluster", map[string]interface{}{"id": d.Id()})

	req := &opensearch.DeleteClusterRequest{
		ClusterId: d.Id(),
	}

	err := waitOperationWithRetry(ctx, config, "Cluster Delete", func() (*operation.Operation, error) {
		return config.sdk.MDB().OpenSearch().Cluster().Delete(ctx, req)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Debug(ctx, "Finished deleting OpenSearch Cluster", map[string]interface{}{"id": d.Id()})

	return nil
}

func resourceYandexMDBOpenSearchClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Debug(ctx, "Updating OpenSearch Cluster", map[string]interface{}{"id": d.Id()})

	d.Partial(true)

	if err := updateOpenSearchClusterParams(ctx, d, meta); err != nil {
		return diag.FromErr(err)
	}

	if err := updateOpenSearchNodeGroupsParams(ctx, d, meta); err != nil {
		return diag.FromErr(err)
	}

	if err := updateDashboardsNodeGroupsParams(ctx, d, meta); err != nil {
		return diag.FromErr(err)
	}

	d.Partial(false)

	tflog.Debug(ctx, "Finishing updating OpenSearch Cluster", map[string]interface{}{"id": d.Id()})

	return resourceYandexMDBOpenSearchClusterRead(ctx, d, meta)
}

func updateOpenSearchNodeGroupsParams(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	oc, nc := d.GetChange("config")
	oldConfig := expandOpenSearchConfigCreateSpec(oc)
	newConfig := expandOpenSearchConfigCreateSpec(nc)
	oldGroups := oldConfig.GetOpensearchSpec().GetNodeGroups()
	newGroups := newConfig.GetOpensearchSpec().GetNodeGroups()

	var oldGroupsByName = map[string]*opensearch.OpenSearchCreateSpec_NodeGroup{}
	for _, g := range oldGroups {
		oldGroupsByName[g.Name] = g
	}
	var newGroupsByName = map[string]*opensearch.OpenSearchCreateSpec_NodeGroup{}
	for _, g := range newGroups {
		newGroupsByName[g.Name] = g
	}

	//Create new nodegroups
	groupsToCreate := make([]*opensearch.OpenSearchCreateSpec_NodeGroup, 0)
	for _, g := range newGroups {
		if _, ok := oldGroupsByName[g.Name]; !ok {
			if dedicatedManagersGroup(g) {
				//add manager group to the start of array
				//add manager group first
				groupsToCreate = append([]*opensearch.OpenSearchCreateSpec_NodeGroup{g}, groupsToCreate...)
			} else {
				groupsToCreate = append(groupsToCreate, g)
			}
		}
	}
	for _, g := range groupsToCreate {
		request, err := createAddOpenSearchNodeGroupRequest(d.Id(), g)
		if err != nil {
			return err
		}
		err = makeAddOpenSearchNodeGroupRequest(ctx, request, d, meta)
		if err != nil {
			return err
		}
	}

	//Update existing nodegroups
	managersToIncrease := make([]*opensearch.UpdateOpenSearchNodeGroupRequest, 0)
	managersToDecrease := make([]*opensearch.UpdateOpenSearchNodeGroupRequest, 0)
	dataManagersToDecrease := make([]*opensearch.UpdateOpenSearchNodeGroupRequest, 0)
	otherGroupsToUpdate := make([]*opensearch.UpdateOpenSearchNodeGroupRequest, 0)
	//to proper update managers count we should use the following sequence:
	//1) Increase hostcount in dedicated manager group if exists
	//2) decrease hostcount in mixed data/manager groups
	//3) do all other operations, including deleting of a group(s)
	//3) decrease hostcount in dedicated manager group if exists
	for _, newGroup := range newGroups {
		if oldGroup, ok := oldGroupsByName[newGroup.Name]; ok {
			request, err := createUpdateOpenSearchNodeGroupRequest(d.Id(), oldGroup, newGroup)
			if len(request.UpdateMask.Paths) == 0 {
				continue
			}
			if err != nil {
				return err
			}
			if dedicatedManagersGroup(newGroup) {
				if newGroup.HostsCount > oldGroup.HostsCount {
					managersToIncrease = append(managersToIncrease, request)
				} else if newGroup.HostsCount < oldGroup.HostsCount {
					managersToDecrease = append(managersToDecrease, request)
				} else {
					otherGroupsToUpdate = append(otherGroupsToUpdate, request)
				}
			} else {
				if (hasManagerRole(newGroup) && newGroup.HostsCount < oldGroup.HostsCount) || (hasManagerRole(oldGroup) && !hasManagerRole(newGroup)) {
					dataManagersToDecrease = append(dataManagersToDecrease, request)
				} else {
					otherGroupsToUpdate = append(otherGroupsToUpdate, request)
				}
			}
		}
	}
	//1) increase managers count
	for _, request := range managersToIncrease {
		err := makeUpdateOpenSearchNodeGroupRequest(ctx, request, d, meta)
		if err != nil {
			return err
		}
	}
	//2) decrease data/managers host count
	for _, request := range dataManagersToDecrease {
		err := makeUpdateOpenSearchNodeGroupRequest(ctx, request, d, meta)
		if err != nil {
			return err
		}
	}

	//3) all other activities
	for _, request := range otherGroupsToUpdate {
		err := makeUpdateOpenSearchNodeGroupRequest(ctx, request, d, meta)
		if err != nil {
			return err
		}
	}

	//Delete old nodegroups
	for _, g := range oldGroups {
		if _, ok := newGroupsByName[g.Name]; !ok {
			request := createDeleteOpenSearchNodeGroupRequest(d.Id(), g)
			err := makeDeleteOpenSearchNodeGroupRequest(ctx, request, d, meta)
			if err != nil {
				return err
			}
		}
	}
	//4) finally decrease host count in managers group
	for _, request := range managersToDecrease {
		err := makeUpdateOpenSearchNodeGroupRequest(ctx, request, d, meta)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateDashboardsNodeGroupsParams(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	oc, nc := d.GetChange("config")
	oldConfig := expandOpenSearchConfigCreateSpec(oc)
	newConfig := expandOpenSearchConfigCreateSpec(nc)
	oldGroups := oldConfig.GetDashboardsSpec().GetNodeGroups()
	newGroups := newConfig.GetDashboardsSpec().GetNodeGroups()

	var oldGroupsByName = map[string]*opensearch.DashboardsCreateSpec_NodeGroup{}
	for _, g := range oldGroups {
		oldGroupsByName[g.Name] = g
	}
	var newGroupsByName = map[string]*opensearch.DashboardsCreateSpec_NodeGroup{}
	for _, g := range newGroups {
		newGroupsByName[g.Name] = g
	}

	//Create new nodegroups
	for _, g := range newGroups {
		if _, ok := oldGroupsByName[g.Name]; !ok {
			request, err := createAddDashboardsNodeGroupRequest(d.Id(), g)
			if err != nil {
				return err
			}
			err = makeAddDashboardsNodeGroupRequest(ctx, request, d, meta)
			if err != nil {
				return err
			}
		}
	}

	//Update existing nodegroups
	for _, newGroup := range newGroups {
		if oldGroup, ok := oldGroupsByName[newGroup.Name]; ok {
			request, err := createUpdateDashboardsNodeGroupRequest(d.Id(), oldGroup, newGroup)
			if len(request.UpdateMask.Paths) == 0 {
				continue
			}
			if err != nil {
				return err
			}
			err = makeUpdateDashboardsNodeGroupRequest(ctx, request, d, meta)
			if err != nil {
				return err
			}
		}
	}

	//Delete old nodegroups
	for _, g := range oldGroups {
		if _, ok := newGroupsByName[g.Name]; !ok {
			request := createDeleteDashboardsNodeGroupRequest(d.Id(), g)
			err := makeDeleteDashboardsNodeGroupRequest(ctx, request, d, meta)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func updateOpenSearchClusterParams(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	req := &opensearch.UpdateClusterRequest{
		ClusterId: d.Id(),
		UpdateMask: &field_mask.FieldMask{
			Paths: make([]string, 0, 16),
		},
	}

	if d.HasChange("description") {
		req.Description = d.Get("description").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "description")
	}

	if d.HasChange("name") {
		req.Name = d.Get("name").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "name")
	}

	if d.HasChange("labels") {
		labelsProp, err := expandLabels(d.Get("labels"))
		if err != nil {
			return err
		}

		req.Labels = labelsProp
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "labels")
	}

	if d.HasChange("config") {
		req.ConfigSpec = expandOpenSearchConfigUpdateSpec(d, req.UpdateMask)
	}

	if d.HasChange("security_group_ids") {
		securityGroupIds := expandSecurityGroupIds(d.Get("security_group_ids"))

		req.SecurityGroupIds = securityGroupIds
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "security_group_ids")
	}

	if d.HasChange("service_account_id") {
		req.ServiceAccountId = d.Get("service_account_id").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "service_account_id")
	}

	if d.HasChange("deletion_protection") {
		req.DeletionProtection = d.Get("deletion_protection").(bool)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "deletion_protection")
	}

	if d.HasChange("maintenance_window") {
		mw, err := expandOpenSearchMaintenanceWindow(d)
		if err != nil {
			return err
		}

		req.MaintenanceWindow = mw
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "maintenance_window")
	}

	if len(req.UpdateMask.Paths) == 0 {
		return nil
	}
	err := makeOpenSearchClusterUpdateRequest(ctx, req, d, meta)
	if err != nil {
		return err
	}

	return nil
}

func makeOpenSearchClusterUpdateRequest(ctx context.Context, req *opensearch.UpdateClusterRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	err := waitOperationWithRetry(ctx, config, "Cluster Update", func() (*operation.Operation, error) {
		return config.sdk.MDB().OpenSearch().Cluster().Update(ctx, req)
	})
	if err != nil {
		return fmt.Errorf("Error updating OpenSearch Cluster %q: %w", d.Id(), err)
	}
	return nil
}

func makeAddOpenSearchNodeGroupRequest(ctx context.Context, req *opensearch.AddOpenSearchNodeGroupRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	err := waitOperationWithRetry(ctx, config, "Add Nodegroup", func() (*operation.Operation, error) {
		return config.sdk.MDB().OpenSearch().Cluster().AddOpenSearchNodeGroup(ctx, req)
	})
	if err != nil {
		return fmt.Errorf("Error updating OpenSearch Cluster (adding nodegroup) %q: %w", d.Id(), err)
	}
	return nil
}

func makeUpdateOpenSearchNodeGroupRequest(ctx context.Context, req *opensearch.UpdateOpenSearchNodeGroupRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	err := waitOperationWithRetry(ctx, config, "Update NodeGroup", func() (*operation.Operation, error) {
		return config.sdk.MDB().OpenSearch().Cluster().UpdateOpenSearchNodeGroup(ctx, req)
	})
	if err != nil {
		return fmt.Errorf("Error updating OpenSearch Cluster (updating nodegroup) %q: %w", d.Id(), err)
	}
	return nil
}

func makeDeleteOpenSearchNodeGroupRequest(ctx context.Context, req *opensearch.DeleteOpenSearchNodeGroupRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	err := waitOperationWithRetry(ctx, config, "Delete NodeGroup", func() (*operation.Operation, error) {
		return config.sdk.MDB().OpenSearch().Cluster().DeleteOpenSearchNodeGroup(ctx, req)
	})
	if err != nil {
		return fmt.Errorf("Error updating OpenSearch Cluster (deleting nodegroup) %q: %w", d.Id(), err)
	}
	return nil
}

func makeAddDashboardsNodeGroupRequest(ctx context.Context, req *opensearch.AddDashboardsNodeGroupRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	err := waitOperationWithRetry(ctx, config, "Add Dashboards NodeGroup", func() (*operation.Operation, error) {
		return config.sdk.MDB().OpenSearch().Cluster().AddDashboardsNodeGroup(ctx, req)
	})
	if err != nil {
		return fmt.Errorf("Error updating OpenSearch Cluster (adding nodegroup) %q: %w", d.Id(), err)
	}
	return nil
}

func makeUpdateDashboardsNodeGroupRequest(ctx context.Context, req *opensearch.UpdateDashboardsNodeGroupRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	err := waitOperationWithRetry(ctx, config, "Update Dashboards NodeGroup", func() (*operation.Operation, error) {
		return config.sdk.MDB().OpenSearch().Cluster().UpdateDashboardsNodeGroup(ctx, req)
	})
	if err != nil {
		return fmt.Errorf("Error updating OpenSearch Cluster (updating dashboards nodegroup) %q: %w", d.Id(), err)
	}
	return nil
}

func makeDeleteDashboardsNodeGroupRequest(ctx context.Context, req *opensearch.DeleteDashboardsNodeGroupRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	err := waitOperationWithRetry(ctx, config, "Delete Dashboards NodeGroup", func() (*operation.Operation, error) {
		return config.sdk.MDB().OpenSearch().Cluster().DeleteDashboardsNodeGroup(ctx, req)
	})
	if err != nil {
		return fmt.Errorf("Error updating OpenSearch Cluster (deleting dashboards nodegroup) %q: %w", d.Id(), err)
	}
	return nil
}

func listOpensearchHosts(ctx context.Context, config *Config, clusterID string) ([]*opensearch.Host, error) {
	hosts := []*opensearch.Host{}
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().OpenSearch().Cluster().ListHosts(ctx, &opensearch.ListClusterHostsRequest{
			ClusterId: clusterID,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error while getting list of hosts for '%s': %s", clusterID, err)
		}
		hosts = append(hosts, resp.Hosts...)
		if resp.NextPageToken == "" || resp.NextPageToken == "0" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return hosts, nil
}

func opensearchNodeGroupsDiffCustomize(ctx context.Context, rdiff *schema.ResourceDiff, meta interface{}) error {
	oc, nc := rdiff.GetChange("config")
	if oc == nil {
		if nc == nil {
			return fmt.Errorf("Missing required option: config")
		}
	}

	var (
		oldConfig = expandOpenSearchConfigCreateSpec(oc)
		newConfig = expandOpenSearchConfigCreateSpec(nc)
	)

	if isNodeGroupsChanged(oldConfig, newConfig) {
		err := rdiff.SetNewComputed("hosts")
		if err != nil {
			return err
		}
	}

	if modifyConfig(oldConfig, newConfig) {
		flattened := flattenOpenSearchConfigCreateSpec(newConfig)
		err := rdiff.SetNew("config", flattened)
		if err != nil {
			return err
		}
	}

	return nil
}

func isNodeGroupsChanged(oldConfig, newConfig *opensearch.ConfigCreateSpec) bool {
	if (oldConfig == nil || newConfig == nil) ||
		(oldConfig.OpensearchSpec == nil || newConfig.OpensearchSpec == nil) ||
		(oldConfig.DashboardsSpec == nil || newConfig.DashboardsSpec == nil) {

		return false
	}

	if len(oldConfig.OpensearchSpec.NodeGroups) != len(newConfig.OpensearchSpec.NodeGroups) {
		return true
	}

	if len(oldConfig.DashboardsSpec.NodeGroups) != len(newConfig.DashboardsSpec.NodeGroups) {
		return true
	}

	if !reflect.DeepEqual(oldConfig.OpensearchSpec.NodeGroups, newConfig.OpensearchSpec.NodeGroups) {
		return true
	}

	if !reflect.DeepEqual(oldConfig.DashboardsSpec.NodeGroups, newConfig.DashboardsSpec.NodeGroups) {
		return true
	}

	return false
}
