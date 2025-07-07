ver=`git rev-parse --short HEAD`


if [ "$(git rev-parse --abbrev-ref HEAD)" = "master" ]; then
    export namespace="imbi-prod"
    export ver="$ver-prod"
    export kube_env="kubernetes"
else
    export namespace="imbi-dev"
    export ver="$ver-dev"
    export kube_env="kubernetes-test"
fi

#deployment to kubernetes (common for all three branch  develop, beta, master)
Repository=gst-apis
project=gst

/usr/bin/sed -i.bak "s/\(^ *image:.*435642640015.*$Repository:\)[^ ]*/\1${ver}/" ../bi-infra/$kube_env/$project/"$project"_deployment.yml 
kubectl  apply -f  ../bi-infra/$kube_env/$project/"$project"_deployment.yml 

#kubectl wait --for=condition=Ready pod -l app=gst-apis -n imbi-dev


# # Check if the pod is running using kubectl wait
# if kubectl wait --for=condition=Ready pod -l app=gst-apis -n $namespace --timeout=720s; then
#   echo "Pod is running successfully."
#   exit 0  # Success, exit with code 0
# else
#   echo "Pod failed to start within the timeout."
#   exit 1  # Failure, exit with code 1
# fi


#curl -H 'Content-Type: application/json' -X POST "" -d "{\"text\":\"$output\n Successfully deployed on $ver \"}" ;

