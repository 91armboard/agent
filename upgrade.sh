#! /bin/sh
# author: zwl
# time: 2018-09-23

upgrade(){
  if [ ! -f "/mnt/mmcblk0p1/ss_main.upgrade" ];then
    echo "please download file first"
  else
    rm /mnt/mmcblk0p1/ss_main
    mv /mnt/mmcblk0p1/ss_main.upgrade /mnt/mmcblk0p1/ss_main
    chmod +x /mnt/mmcblk0p1/ss_main
    killall ss_main
    sh /root/agent.sh
  fi
}

upgrade
