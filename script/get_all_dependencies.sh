#!/bin/bash

# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

location=$(dirname $0)
rootdir=$location/../

if [ "$#" -lt 1 ]; then
  echo "usage: $0 <Camel K runtime version> [<staging repository>]"
  exit 1
fi
runtime_version="$1"

if [ ! -z $2 ]; then
  # Change the settings to include the staging repo if it's not already there
  echo "INFO: updating the settings staging repository"
  sed -i.bak "s;<url>https://repository\.apache\.org/content/repositories/orgapachecamel-.*</url>;<url>$2</url>;" $location/maven-settings.xml
  rm $location/maven-settings.xml.bak
fi

catalog="$rootdir/resources/camel-catalog-$runtime_version.yaml"
ckr_version=$(yq .spec.runtime.version $catalog)
cq_version=$(yq '.spec.runtime.metadata."camel-quarkus.version"' $catalog)

sed 's/- //g' $catalog | grep "groupId\|artifactId" | paste -d " "  - - | awk '{print $2,":",$4}' | tr -d " " | sort | uniq > /tmp/ck.dependencies

dependencies=$(cat /tmp/ck.dependencies)
for d in $dependencies
do
    mvn_dep=""
    if [[ $d == org.apache.camel.quarkus* ]]; then
        mvn_dep="$d:$cq_version"
    elif [[ $d == org.apache.camel.k* ]]; then
        mvn_dep="$d:$ckr_version"
        # TODO merge with package_maven_artifacts.sh script
        continue
    else
        echo "ERROR: cannot parse $d kind of dependency"
        exit 1
    fi
    echo "INFO: downloading $mvn_dep and its transitive dependencies..."
    mvn -q dependency:get -Dartifact=$mvn_dep -Dmaven.repo.local=${rootdir}build/_maven_output -s $location/maven-settings.xml
done