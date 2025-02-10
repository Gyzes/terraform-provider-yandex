---
subcategory: "Lockbox (Secret Management)"
page_title: "Yandex: yandex_lockbox_secret_version_hashed"
description: |-
  Manages Yandex Cloud Lockbox secret version (with values hashed in state).
---

# yandex_lockbox_secret_version_hashed (Resource)

Yandex Cloud Lockbox secret version resource (with values hashed in state). For more information, see [the official documentation](https://yandex.cloud/docs/lockbox/).

## Example usage

```terraform
//
// Create a new Lockbox Secret Hashed Version.
//
resource "yandex_lockbox_secret" "my_secret" {
  name = "test secret"
}

resource "yandex_lockbox_secret_version_hashed" "my_version" {
  secret_id = yandex_lockbox_secret.my_secret.id
  key_1     = "key1"
  // in Terraform state, these values will be stored in hash format
  text_value_1 = "sensitive value 1"
  key_2        = "k2"
  text_value_2 = "sensitive value 2"
  // etc. (up to 10 entries)
}
```

## Argument Reference

The following arguments are supported:

* `secret_id` - (Required) The Yandex Cloud Lockbox secret ID where to add the version.
* `description` - (Optional) The Yandex Cloud Lockbox secret version description.
* `key_<NUMBER>` - (Optional) Each of the entry keys in the Yandex Cloud Lockbox secret version.
* `text_value_<NUMBER>` - (Optional) Each of the entry values in the Yandex Cloud Lockbox secret version.

The `<NUMBER>` can range from `1` to `10`. If you only need one entry, use `key_1`/`text_value_1`. If you need a second entry, use `key_2`/`text_value_2`, and so on.


## Import

~> Import for this resource is not implemented yet.

