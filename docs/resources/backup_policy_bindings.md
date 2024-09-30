---
subcategory: "Backup"
page_title: "Yandex: yandex_backup_policy_bindings"
description: |-
  Allows to bind compute instance with backup policy.
---


# yandex_backup_policy_bindings




Allows management of [Yandex.Cloud Attach and Detach VM](https://yandex.cloud/docs/backup/operations/policy-vm/attach-and-detach-vm).

> **NOTE:\_** Cloud Backup Provider must be activated in order to manipulate with policies. Active it either by UI Console or by `yc` command.

```terraform
resource "yandex_iam_service_account" "test_sa" {
  name = "sa-backup-editor"
}

resource "yandex_resourcemanager_folder_iam_member" "test_binding" {
  folder_id = yandex_iam_service_account.test_sa.folder_id
  role      = "backup.editor"
  member    = "serviceAccount:${yandex_iam_service_account.test_sa.id}"
}

resource "yandex_vpc_network" "test_backup_network" {}

resource "yandex_vpc_subnet" "test_backup_subnet" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.test_backup_network.id
  v4_cidr_blocks = ["192.168.0.0/24"]
}

# https://yandex.cloud/ru/docs/backup/concepts/vm-connection#vm-network-access
resource "yandex_vpc_security_group" "test_backup_security_group" {
  name       = "cloud-backup"
  network_id = yandex_vpc_network.test_backup_network.id
  egress {
    protocol       = "TCP"
    from_port      = 7770
    to_port        = 7800
    v4_cidr_blocks = ["84.47.172.0/24"]
  }
  egress {
    protocol       = "TCP"
    port           = 443
    v4_cidr_blocks = ["213.180.204.0/24", "213.180.193.0/24", "178.176.128.0/24", "84.201.181.0/24", "84.47.172.0/24"]
  }
  egress {
    protocol       = "TCP"
    port           = 80
    v4_cidr_blocks = ["213.180.204.0/24", "213.180.193.0/24"]
  }
  egress {
    protocol       = "TCP"
    port           = 8443
    v4_cidr_blocks = ["84.47.172.0/24"]
  }
  egress {
    protocol       = "TCP"
    port           = 44445
    v4_cidr_blocks = ["51.250.1.0/24"]
  }
}

data "yandex_compute_image" "ubuntu" {
  family = "ubuntu-2004-lts"
}

resource "yandex_compute_instance" "test_backup_compute" {
  name        = "test-backup-compute"
  platform_id = "standard-v1"
  zone        = "ru-central1-a"

  service_account_id = yandex_iam_service_account.test_sa.id

  network_interface {
    subnet_id          = yandex_vpc_subnet.test_backup_subnet.id
    security_group_ids = [yandex_vpc_security_group.test_backup_security_group.id]
    nat                = true
  }

  boot_disk {
    initialize_params {
      image_id = data.yandex_compute_image.ubuntu.id
    }
  }

  resources {
    cores  = 2
    memory = 4
  }

  metadata = {
    user-data = "#cloud-config\npackages:\n  - curl\n  - perl\n  - jq\nruncmd:\n  - curl https://storage.yandexcloud.net/backup-distributions/agent_installer.sh | sudo bash\n"
  }
}

data "yandex_backup_policy" "test_backup_policy" {
  name = "Default daily"
}

resource "yandex_backup_policy_bindings" "test_backup_binding" {
  instance_id = yandex_compute_instance.test_backup_compute.id
  policy_id   = data.yandex_backup_policy.test_backup_policy.id
}
```

## Argument Reference

The following arguments are supported:

- `instance_id` (**Required**) — Compute Cloud instance ID.
- `policy_id` (**Required**) — Backup Policy ID.

## Attributes Reference

The following attributes are exported:

* `created_at` (Computed) - Creation timestamp of the Yandex Cloud Policy Bindings.
* `processing` (Computed) - Boolean flag that specifies whether the policy is in the process of binding to an instance.
* `enabled` (Computed) - Boolean flag that specifies whether the policy application is enabled. May be `false` if `processing` flag is `true`.
