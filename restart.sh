#! /bin/sh
# author: zwl
# time: 2018-09-23

restart(){
  killall ss_main
  sh /root/agent.sh
}

restart
