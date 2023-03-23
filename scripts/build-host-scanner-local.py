"""
Script providing a way to test the Host-Scanne on a local env. with private K8s env.
It is installing the dependencies and initializing the relevant k8s env.
If any issues or error ocurred during the run, script will print it out. 
after the script completed, you can perform any chnage at the host-scanner and test it with kubescape run. 

Help: 
build-host-scanner-local.py --build - building the env. and preparing for local run 
build-host-scanner-local.py --revert - destroying the env. and closing local run 
build-host-scanner-local.py --help - showing help

Dependencies: 
1. minikube is installed
2. docker is enabled in the system 
3. kubectl is installed
4. for cloud providers:
    Access to a remote private repository such as dockerhub.
    Access to a cloud provider running cluster.
"""
    
import subprocess
from datetime import datetime 
import sys
import os
import logging

# global parameters defenition
CURRENTIMAGENAME = 'quay.io/kubescape/host-scanner'
CURRENTTAG = 'latest'
HOSTSCANNERIMAGENAME = 'quay.io/kubescape/host-scanner'
TMPIMAGENAME = 'local-host-scanner-image'
DEPLOYMENTYAML = os.getcwd() + '/deployment/k8s-deployment.yaml'
DOCKERFILE = os.getcwd() + '/build/Dockerfile'


"""
fucnction is running any OS command and can return it's output/error.
:param `cmd` represent the command to run as a string
:param `format` represent the output format (utf / json)
:param `resFormat` represent the return format (int/str)
:return command console output if succeeded, otherwise return (-1) 
"""
def runCmdCommand(cmd, resFormat):

    try:
        logging.info(f"Running command: '{cmd}'")

        proc = subprocess.run(cmd, shell=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
        lines = proc.stdout.decode('utf-8').strip()
        err = proc.stderr
        
        if err != None:
            logging.error(f"Operation finished with error code '{proc.returncode}' with error '{err}'")
            return -1
        
        if resFormat == "int":
            return 0
        else: 
            return lines
    except subprocess.CalledProcessError as e:
        logging.error(f"Command returned error '{e.output}' and error code '{e.returncode}'")
        return -1

"""
function is updating k8s deployment file
:param `flagType` define if we need to create unique image abd tag name or revert to original (revert/deploy)
"""
def updateK8sDeploymentFile(flagType):
    
    logging.info(f"{flagType}ing K8s file")
    
    bool = False
    imagePullPolicyBool = False
    
    try:
        # open k8s-deployment file and search for host-scanner image
        with open (DEPLOYMENTYAML,'r') as f:
            lines = f.readlines()
            for i, line in enumerate(lines):
                    if flagType == 'deploy':
                        # verify that original host-scanner image is configured
                        if line.__contains__(HOSTSCANNERIMAGENAME):
                            line = line.replace(HOSTSCANNERIMAGENAME, TMPIMAGENAME).replace('latest', 'test')
                            bool = True   # check that correct image found. 
                            lines[i] = line
                            lines.insert(i+1, '        imagePullPolicy: Never\n')
                    # reverting image to original host-scanner image name
                    if flagType == 'revert' and line.__contains__(TMPIMAGENAME):
                        line = line.replace(TMPIMAGENAME, HOSTSCANNERIMAGENAME).replace('test', 'latest')
                        bool = True   # check that correct image found. 
                        lines[i] = line
                    if flagType == 'revert' and line.__contains__('imagePullPolicy'):
                        del lines[i]
                        imagePullPolicyBool = True   # check imagePullPolicy found. 
            f.close
        if bool == False:
            logging.error("Couldn't find the right Host-Scanner image name or object.\
                  \nCheck the value under: k8s-deployment.yaml --> spec:template:spec:containers:image\n")
            return -1  
        if imagePullPolicyBool == False and flagType == 'revert':
            logging.error("Couldn't find the right Host-Scanner imagePullPolicy object.\nCheck the value under: k8s-deployment.yaml --> spec:template:spec:containers:imagePullPolicy")
            return -1  
        else:
            # write changes to doc
            try:
                with open (DEPLOYMENTYAML,'w') as f:
                    for i in lines:
                        f.write(i)
                f.close
                return 0
            except Exception as e:
                logging.error(e)
                return -1
    except Exception as e:
        logging.error(e)
        return -1

""" 
function is configuring the necessary GO env. 
return 0 for success / otherwise -1
"""
def configGoEnv():
    
    logging.info("Configuring GO environmnet")

    goPath = runCmdCommand('which go', 'str')
    if goPath == -1:
        return -1

    # update GOPATH
    cmd = 'export GOPATH="' + goPath + '"'
    if runCmdCommand(cmd, 'int') == -1:
        return -1
    
    # update PATH with latest GOPATH
    if runCmdCommand('export PATH=$PATH:$GOPATH/bin', 'int') == -1:
        return -1
    
    # install kubectl-curl if not installed 
    if runCmdCommand('go install github.com/segmentio/kubectl-curl@latest', 'int'):
        return -1

    return 0


"""
function reads the input arguments and initialize the processes
:return 0 build success / -1 build failed 
"""
def readArgs(args):
    # print help and usage
    if '--help' in args:
        logging.info("""
                Script providing a way to test the Host-Scanne on a local env. with private K8s env.
                It is installing the dependencies and initializing the relevant k8s env.
                If any issues or error ocurred during the run, script will print it out. 
                after the script completed, you can perform any chnage at the host-scanner and test it with kubescape run. 

                Help: 
                build-host-scanner-local.py --build - building the env. and preparing for local run 
                build-host-scanner-local.py --revert - destroying the env. and closing local run 
                build-host-scanner-local.py --help - showing help

                Dependencies: 
                1. minikube is installed
                2. docker is enabled in the system 
                3. kubectl is installed
                4. for cloud providers:
                    Access to a remote private repository such as dockerhub.
                    Access to a cloud provider running cluster.
                """)

    
    # reverting the local env. and remove it
    if '--revert' in args:
        logging.info("Reverting")

        # deploy chnages to k8s deployment file 0 - success / -1 failure
        if updateK8sDeploymentFile('revert') == -1:
            return -1

        # apply file changes to k8s
        cmd = 'kubectl apply -f ' + DEPLOYMENTYAML
        res = runCmdCommand(cmd, 'str')   
        if res == -1:    
            return -1
        else:
            logging.info(res)
        
        # build the old docker image 
        cmd = 'docker build -f ' + DOCKERFILE + ' . -t ' + HOSTSCANNERIMAGENAME + ':latest'
        res = runCmdCommand(cmd, 'str')
        if res == -1:    
            return -1
        else:
            logging.info(res)
        
        # kill minikube 
        res = runCmdCommand('minikube stop', 'str') 
        if res == -1:    
            return -1
        else:
            logging.info(res)
    
    # building the local env. and setup it    
    if '--build' in args:
        logging.info("Building Host-Scanner localy")

        # config GO env. 0-success / -1-failure
        if configGoEnv() == -1:
            return -1
        
        # init minikube
        res = runCmdCommand('minikube start', 'str')
        if res == -1:    
            return -1
        else:
            logging.info(res)
        
        # deploy changes to k8s deployment file
        if updateK8sDeploymentFile('deploy') == -1:
            return -1
        
        # apply file changes to k8s
        cmd = 'kubectl apply -f ' + DEPLOYMENTYAML
        res = runCmdCommand(cmd, 'str')   
        if res == -1:    
            return -1
        else:
            logging.info(res)

        # Point terminal’s docker-cli to the Docker Engine inside minikube:
        res = runCmdCommand('eval $(minikube docker-env)', 'str') 
        if res == -1:    
            return -1
        else:
            logging.info(res)
        
        # build the docker image 
        cmd = 'docker build -f ' + DOCKERFILE + ' . -t ' + TMPIMAGENAME + ':test'
        res = runCmdCommand(cmd, 'str')
        if res == -1:    
            return -1
        else:
            logging.info(res)
            
    return 0


if __name__ == '__main__': 
    
    # TODO: add support with cloud providers
    # TODO: add support with more than 1node for testing
    
    logging.basicConfig(format='%(asctime)s | %(levelname)s | %(message)s', datefmt='%d-%m-%YT%H:%M:%S', level=logging.INFO)
    logging.info("Armo's Host-Scanner local environment build strated")

    args = sys.argv
    res = readArgs(args)

    if res == -1:
        logging.error("Armo's Host-Scanner local environment build failed")
    else:
        logging.info("Armo's Host-Scanner local environment build ended successfully")

    sys.exit(res)