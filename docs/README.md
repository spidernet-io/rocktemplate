# Introduction

## copy

1. copy repo

   replace all 'rocktemplate' to 'YourRepoName'

   replace all 'spidernet-io' and 'spidernet.io' to 'YourOrigin'

2. grep "====modify====" * -RHn --colour  and modify all of them

3. update api/v1/openapi.yaml and `make update_openapi_sdk`

4. redefine CRD in pkg/k8s/v1
    rename directory name 'pkg/k8s/v1/rocktemplate.spidernet.io' 
    replace all 'mybook' to 'YourCRDName'
    and `make update_crd_sdk`, and code pkg/mybookManager

    rename pkg/mybookManager and replace all 'mybook' with your CRD name in this directory

    rm charts/crds/rocktemplate.spidernet.io_mybooks.yaml 

    in repo: replace all "github.com/spidernet-io/spiderdoctor/pkg/mybookManager" to "github.com/spidernet-io/spiderdoctor/pkg/${crdName}Manager"
    in repo: find and replace all "mybook" to YourCrd

5. update charts/ , and images/ , and CODEOWNERS

6. `go mod tidy` , `go mod vendor` , `go vet ./...` , double check all is ok

7. `go get -u` , `go mod tidy` , `go mod vendor` , `go vet ./...`  , update all vendor

8. create an empty branch 'github_pages' and mkdir 'docs'

9. github seetings:

   spidernet.io  -> settings -> secrets -> actions -> grant secret to repo

   spidernet.io/REPO  -> settings -> general -> feature -> issue

   spidernet.io/ORG  -> settings -> actions -> general -> allow github action to create pr
   spidernet.io/REPO  -> settings -> actions -> general -> allow github action to create pr

   spidernet.io  -> settings -> packages -> public 

   repo -> packages -> package settings -> Change package visibility

   create 'github_pages' branch, and repo -> settings -> pages -> add branch 'github_pages', directory 'docs'

   repo -> settings -> branch -> add protection rules for 'main' and 'github_pages' and 'release*'

   repo -> settings -> tag -> add protection rules for tags

10. enable third app

    personal github -> settings -> applications -> configure

    codefactor: https://www.codefactor.io/dashboard

    sonarCloud: https://sonarcloud.io/projects/create

    codecov: https://app.codecov.io/gh

11. add badge to readme:

    github/workflows/call-e2e.yaml

    github/workflows/badge.yaml

    auto nightly ci

    release version

    code coverage from https://app.codecov.io/gh

    go report from https://goreportcard.com

    codefactor: https://www.codefactor.io/dashboard

    sonarCloud: https://sonarcloud.io/projects

12. build base image , 
    update BASE_IMAGE in images/agent/Dockerfile and images/controller/Dockerfile
    run test
 


## local develop

1. `make build_local_image`

2. `make e2e_init`

3. `make e2e_run`

4. check proscope, browser vists http://NodeIP:4040

5. apply cr

        cat <<EOF > mybook.yaml
        apiVersion: rocktemplate.spidernet.io/v1
        kind: Mybook
        metadata:
          name: test
        spec:
          ipVersion: 4
          subnet: "1.0.0.0/8"
        EOF
        kubectl apply -f mybook.yaml

## chart develop

helm repo add rock https://spidernet-io.github.io/rocktemplate/

