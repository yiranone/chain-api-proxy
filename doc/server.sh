os=`uname -s`
prog="chain-api-proxy"
DEPLOY_HOME=/home/server/chain-api-proxy
DEPLOY_BIN=$DEPLOY_HOME/$prog

echo "enter dir: $DEPLOY_HOME"
cd $DEPLOY_HOME

check_kill(){
    pid=$1
    timeout=15
    while [ $timeout -gt 1 ]
    do
        `ps -eo pid | grep -w $pid >/dev/null`
        if [ $? -eq 0 ]
    then
        sleep 1
        timeout=`expr $timeout - 1`
        else
        echo "process $pid exit successfully"
        return 0
        fi
    done
    echo 'force exiting...'
    kill -9 $pid
    sleep 5
}

linux_start(){
    echo "Starting... $prog ";
    echo "配置文件$DEPLOY_HOME/config.yaml"
    setsid $DEPLOY_BIN --config=$DEPLOY_HOME/config.yaml 1>>$DEPLOY_HOME/logs/boot.log 2>&1 &
    echo "success started $prog ";
}

linux_stop(){
    echo "Stopping... $prog "
    ppid=`ps -eo pid,ppid,command | grep $prog | awk -F' ' '$2==1 {print $1}'`

    if [ "x$ppid" = "x" -o "x$ppid" = "x1" -o "x$ppid" = "x0" ];then
        echo " $prog not running."
    else
        #先尝试优雅退出
        for pid in $ppid
        do
            echo "先尝试优雅退出 kill -s SIGINT $pid"
            kill $pid
            echo "sleep 10"
            sleep 10
            echo "check_kill $pid"
            check_kill $pid
        done
    fi
    echo "stop $prog finish"
}

linux_control(){
    if test $status = "stop"
    then
        linux_stop
    fi

    if test $status = "start"
    then
        linux_start
    fi

    if test $status = "restart"
    then
        linux_stop
        sleep 3
        linux_start
    fi
}

if [ $# -lt 1 ]; then
    echo "$0 start, stop, restart"
    exit 3
fi


status=$1
if test -z $status
then
    echo "$0 start, stop, restart"
    exit 3
fi

if test $status != "start" -a $status != "stop" -a $status != "restart"
then
    echo "$0 start, stop, restart"
    exit 3
fi

if test $os = "Linux"
then
    linux_control
else
    echo "no support $os"
fi

exit 0