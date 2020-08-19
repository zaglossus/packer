# This file was autogenerate by the BETA 'packer hcl2_upgrade' command. We
# recommend double checking that everything is correct before going forward. We
# also recommend treating this file as disposable. The HCL2 blocks in this
# file can be moved to other files. For example, the variable blocks could be
# moved to their own 'variables.pkr.hcl' file, etc. Those files need to be
# suffixed with '.pkr.hcl' to be visible to Packer. To use multiple files at
# once they also need to be in the same folder. 'packer inspect folder/'
# will describe to you what is in that folder.

# All generated input variables will be of string type as this how Packer JSON
# views them; you can later on change their type. Read the variables type
# constraints documentation
# https://www.packer.io/docs/from-1.5/variables#type-constraints for more info.
variable "aws_access_key" {
  type      = string
  default   = "the_default_key"
  sensitive = true
}

variable "aws_region" {
  type    = string
  default = "eu-west-1"
}

variable "aws_secret_key" {
  type      = string
  sensitive = true
}

# "timestamp" template function replacement
locals { timestamp = regex_replace(timestamp(), "[- TZ:]", "") }

# source blocks are generated from your builders; a source can be referenced in
# build blocks. A build block runs provisioner and post-processors onto a
# source. Read the documentation for source blocks here:
# https://www.packer.io/docs/from-1.5/blocks/source
source "amazon-ebs" "autogenerated_1" {
  access_key      = "${var.aws_access_key}"
  ami_description = "Ubuntu 16.04 LTS - expand root partition"
  ami_name        = "ubuntu-16.04 test ${local.timestamp}"
  encrypt_boot    = true
  launch_block_device_mappings {
    delete_on_termination = true
    device_name           = "/dev/sda1"
    volume_size           = 48
    volume_type           = "gp2"
  }
  launch_block_device_mappings {
    delete_on_termination = false
    device_name           = "/dev/sda2"
    volume_size           = 42
    volume_type           = "gp2"
  }
  region     = "${var.aws_region}"
  secret_key = "${var.aws_secret_key}"
  source_ami_filter {
    filters     = { name = "ubuntu/images/*/ubuntu-xenial-16.04-amd64-server-*", root-device-type = "ebs", virtualization-type = "hvm" }
    most_recent = true
    owners      = ["099720109477"]
  }
  spot_instance_types = ["t2.small", "t2.medium", "t2.large"]
  spot_price          = "0.0075"
  ssh_username        = "ubuntu"
}

# a build block invokes sources and runs provisionning steps on them. The
# documentation for build blocks can be found here:
# https://www.packer.io/docs/from-1.5/blocks/build
build {
  sources = ["source.amazon-ebs.autogenerated_1"]

  provisioner "shell" {
    inline      = ["echo ${var.secret_account}", "echo ${build.ID}", "echo ${build.SSHPrivateKey}", "sleep 100000"]
    max_retries = 5
    only        = ["builder"]
  }
  provisioner "shell-local" {
    except  = ["other_builder"]
    inline  = ["sleep 100000"]
    timeout = "5s"
  }
  post-processor "amazon-import" {
    format         = "vmdk"
    license_type   = "BYOL"
    region         = "eu-west-3"
    s3_bucket_name = "hashicorp.adrien"
    tags           = { Description = "packer amazon-import ${local.timestamp}" }
  }
  post-processors {
    post-processor "artifice" {
      keep_input_artifact = true
      files               = ["path/something.ova"]
      name                = "very_special_artifice_post-processor"
      only                = ["builder"]
    }
    post-processor "amazon-import" {
      except         = ["other_builder"]
      license_type   = "BYOL"
      s3_bucket_name = "hashicorp.adrien"
      tags           = { Description = "packer amazon-import ${local.timestamp}" }
    }
  }
}
