package nsentry

/*
#include <stdlib.h>
#include<errno.h>
#include<sched.h>
#include<stdio.h>
#include<stdlib.h>
#include<string.h>
#include<fcntl.h>

void __attribute__((constructor)) enter_namespace(void){
    char *mydocker_pid;
    // 从环境变量中获取需要进入的PID
    mydocker_pid = getenv("mydocker_pid");
    if (!mydocker_pid){
        //fprintf(stdout,"messing mydocker_pid=%s\n",mydocker_pid);
        return;
    }
    // 从环境变量中获取cmd
    char *docker_cmd;
    docker_cmd = getenv("mydocker_cmd");
    if (!docker_cmd){
        //fprintf(stdout,"messing mydocker_cmd=%s\n",docker_cmd);
        return;
    }
    int i = 0;
    char nspath[1024];
    //要进入的五个namespace
    char *namespaces[] = {"ipc","uts","net","pid","mnt"};
    for (i = 0; i < 5; i++){
        //拼接ns的路径
        sprintf(nspath,"/proc/%s/ns/%s",mydocker_pid,namespaces[i]);
        //printf(nspath);
        int fd = open(nspath,O_RDONLY);
        //调用setns进入ns
        if(setns(fd,0 == -1)){
        //    fprintf(stderr,"setns on %s namespace failed: %s \n",namespaces[i],strerror(errno));
            return;
        }
        close(fd);
    }
    // 在进入的Namespace中执行的命令
    int res = system(docker_cmd);
    exit(0);
    return;
}

*/
import "C"
