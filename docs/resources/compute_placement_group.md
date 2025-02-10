---
subcategory: "Compute Cloud"
page_title: "Yandex: yandex_compute_placement_group"
description: |-
  Manages a Placement group resource.
---

# yandex_compute_placement_group (Resource)

A Placement group resource. For more information, see [the official documentation](https://yandex.cloud/docs/compute/concepts/placement-groups).

## Example usage

```terraform
//
// Create a new Compute Placement Group.
//
resource "yandex_compute_placement_group" "group1" {
  name        = "test-pg"
  folder_id   = "abc*********123"
  description = "my description"
}
```

## Argument Reference

The following arguments are supported:

* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.

* `name` - (Optional) The name of the Placement Group.

* `description` - (Optional) A description of the Placement Group.

* `labels` - (Optional) A set of key/value label pairs to assign to the Placement Group.

* `placement_strategy_spread` - A placement strategy with spread policy of the Placement Group. Should be true or unset (conflicts with placement_strategy_partitions).

* `placement_strategy_partitions` - A number of partitions in the placement strategy with partitions policy of the Placement Group (conflicts with placement_strategy_spread).

## Timeouts

This resource provides the following configuration options for [timeouts](https://www.terraform.io/docs/language/resources/syntax.html#operation-timeouts):

- `create` - Default is 1 minute.
- `update` - Default is 1 minute.
- `delete` - Default is 1 minute.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```bash
# terraform import yandex_compute_placement_group.<resource Name> <resource Id>
terraform import yandex_compute_placement_group.my_pg1 ...
```
