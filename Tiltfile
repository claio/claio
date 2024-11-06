load('ext://restart_process', 'docker_build_with_restart')

DOCKERFILE = '''FROM golang:alpine
RUN mkdir /app
RUN chown 65532:65532 /app
USER 65532:65532
WORKDIR /app
COPY ./tilt_bin/manager /app
CMD ["/app/manager"]
'''

def manifests():
    return 'controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases;'

def generate():
    return 'controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./...";'

def compile():
    return 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o tilt_bin/manager cmd/main.go;'

local(generate() + manifests())

local_resource( 'CRDs', 
                manifests() + 'kubectl apply -k config/crd', 
                deps=['api'],
                labels=["controller"])

local_resource( 'Compile', 
                generate() + compile(),                
                deps=['internal', 'cmd/main.go', 'api'],
                ignore=['*/*/zz_generated.deepcopy.go'],
                labels=["controller"])

k8s_yaml(local('kustomize build config/default'))
k8s_resource('claio-controller-manager', labels=['controller'])

docker_build_with_restart('controller:latest', '.', 
    dockerfile_contents=DOCKERFILE,
    entrypoint='/app/manager',
    only=['./tilt_bin/manager'],
    live_update=[
        sync('./tilt_bin/manager', '/app/manager'),
    ]
)

include('./kine/Tiltfile')
include('./test/tenant-sample/Tiltfile')
