#!/bin/bash

 if [ x"${REGISTRY}" == "x" ]; then
     echo "Build version $BUILD_VERSION push to dockerhub"
     docker  build  .  -t suikast42/logunifier:$BUILD_VERSION
     echo docker push suikast42/logunifier:$BUILD_VERSION
     docker push suikast42/logunifier:$BUILD_VERSION
 else
     echo "Build version $BUILD_VERSION push to $REGISTRY"
     docker  build  .  -t  $REGISTRY/suikast42/logunifier:$BUILD_VERSION
     echo docker push $REGISTRY/suikast42/logunifier:$BUILD_VERSION
     docker push $REGISTRY/suikast42/logunifier:$BUILD_VERSION
 fi
