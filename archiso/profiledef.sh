#!/usr/bin/env bash
# shellcheck disable=SC2034

iso_name="yorha-linux"
iso_label="YORHA_$(date --date="@${SOURCE_DATE_EPOCH:-$(date +%s)}" +%Y%m)"
iso_publisher="Lewis Cook <https://www.lcook.net>"
iso_application="YoRHa Linux Installer"
iso_version="$(date --date="@${SOURCE_DATE_EPOCH:-$(date +%s)}" +%Y.%m.%d)"
install_dir="arch"
buildmodes=('iso')
bootmodes=('uefi.grub')
pacman_conf="pacman.conf"
airootfs_image_type="erofs"
airootfs_image_tool_options=('-zlzma,109' -E 'ztailpacking')
bootstrap_tarball_compression=(xz -9e)
file_permissions=(
  ["/etc/shadow"]="0:0:400"
  ["/root"]="0:0:755"
  ["/root/.bash_profile"]="0:0:755"
  ["/usr/local/bin/yorha-inst"]="0:0:755"
)
