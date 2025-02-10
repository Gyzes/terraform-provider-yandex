---
subcategory: "Cloud Organization"
page_title: "Yandex: yandex_organizationmanager_user_ssh_key"
description: |-
  Get information about a Yandex Cloud User SSH Key.
---

# yandex_organizationmanager_user_ssh_key (Data Source)

Get information about a Yandex Cloud User SSH Key.

## Example usage

```terraform
//
// Get information about existing OrganizationManager User SSH Key.
//
data "yandex_organizationmanager_user_ssh_key" "my_user_ssh_key" {
  user_ssh_key_id = "some_user_ssh_key_id"
}

output "my_user_ssh_key_name" {
  value = "data.yandex_organizationmanager_user_ssh_key.my_user_ssh_key.name"
}
```

## Argument Reference

* `user_ssh_key_id` - (Required) ID of the user ssh key.

## Attributes Reference

The following attributes are exported:

* `organization_id` - Organization that the user ssh key belongs to.
* `subject_id` - Subject that the user ssh key belongs to.
* `name` - Name of the user ssh key.
* `data` - Data of the user ssh key.
* `fingerprint` - Auto generated fingerprint of the user ssh key.
* `created_at` - User ssh key creation timestamp.
* `expires_at` - User ssh key will be no longer valid after expiration timestamp.
