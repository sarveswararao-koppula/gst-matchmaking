# this file builds a docker container for production env with current code
# and pushed the image to the registry


git rev-parse HEAD > ./REVISION
# ver=`git describe --tags --abbrev=0`
ver=`git rev-parse --short HEAD`
RELEASE=`git show --stat --pretty=short --abbrev=8 HEAD`


if [ "$(git rev-parse --abbrev-ref HEAD)" = "master" ]; then
    export ver="$ver-prod"
else  
    export ver="$ver-dev"
fi

docker build --platform linux/amd64 -t gst-apis:latest -f Dockerfile .

docker tag gst-apis:latest gst-apis:$ver
docker tag gst-apis:$ver 435642640015.dkr.ecr.ap-south-1.amazonaws.com/gst-apis:$ver
# get login credentials for aws ecr below
aws ecr get-login-password --region ap-south-1 | docker login --username AWS --password-stdin 435642640015.dkr.ecr.ap-south-1.amazonaws.com
# push the image to AWS ECR
docker push 435642640015.dkr.ecr.ap-south-1.amazonaws.com/gst-apis:$ver

