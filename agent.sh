#! /bin/sh
# author: zwl
# time: 2018-09-23 

function mount_data(){
    echo "mount_data"
    mount_count=`df | grep /dev/sda1 | grep /data | wc -l`
    if [ 1 == $mount_count ]; then
        ## Mounted
        echo "Mounted"
    else
        mount /dev/sda1 /data
        sleep 5s
    fi
}

function start_agent(){
    echo "start_agent"
    main_count=`ps | grep /data/main | wc -l`
    if [ 1 == $main_count ]; then
        main_file_count=`ls /data | grep ^main$ | wc -l`
        if [ 1 == $main_file_count ]; then
            /data/main > /data/main.log 2>&1 &
        fi
    fi

    frpc_count=`ps | grep /data/frpc | wc -l`
    if [ 1 == $frpc_count ]; then
        frpc_file_count=`ls /data | grep ^frpc$ | wc -l`
        if [ 1 == $frpc_file_count ]; then
            /data/frpc -c /data/config/http.ini > /data/frp.log 2>&1  &
        fi
    fi
}
mount_data
start_agent