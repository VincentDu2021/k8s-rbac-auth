#!/bin/bash

#KUBECONFIG need to be set to cluster-admin
list=($(kubectl get --raw / | jq -r '.paths[] | select(. | startswith("/api"))'))
#NAME                              SHORTNAMES            APIVERSION                             NAMESPACED   KIND                             VERBS

printf '%-38s%-22s%-39s%-13s%-33s%-60s\n' "NAME" "SHORTNAMES" "APIVERSION" "NAMESPACED" "KIND" "VERBS"
for ep in ${list[@]}; do
    res=$(kubectl get --raw ${ep} | jq .resources)
    if [ "x${res}" != "xnull" ]; then
        #    kubectl get --raw ${ep} | jq -r ".resources[] | .name,.verbs";
        APIGROUPVERSION=$(kubectl get --raw ${ep} | jq -r .groupVersion)
        RESOURCES=$(kubectl get --raw ${ep} | jq -r .resources[])
        NAMES=($(echo $RESOURCES | jq -r .name))
        NAMESPACED=($(echo $RESOURCES | jq -r .namespaced))
        KINDS=($(echo $RESOURCES | jq -r .kind))

        temp=($(echo $RESOURCES | jq -r .shortNames))
        SHORTNAMES=()
        for s in ${temp[*]}; do
            if [[ $s == null ]]; then
                SHORTNAMES+=("[]")
            else
                t+=$s
                if [[ $s == ']' ]]; then
                    SHORTNAMES+=($t)
                    t=""
                fi
            fi
        done

        temp=($(echo $RESOURCES | jq -r .verbs))
        VERBS=()
        for s in ${temp[*]}; do
            if [[ $s == null ]]; then
                VERBS+=("[]")
            else
                t+=$s
                if [[ $s == ']' ]]; then
                    VERBS+=($t)
                    t=""
                fi
            fi
        done
        LEN=$((${#NAMES[@]} - 1))

        for i in $(seq 0 $LEN); do
            if [[ null == ${SHORTNAMES[$i]} ]]; then
                SN=""
            else
                SN=($(echo "${SHORTNAMES[$i]}" | jq -r '.| flatten[]'))
                SN=$(printf '%s\n' "$(
                    IFS=,
                    printf '%s' "${SN[*]}"
                )")
            fi

            if [[ null == ${VERBS[$i]} ]]; then
                VB="[]"
            else
                VB=($(echo "${VERBS[$i]}" | jq -r '.| flatten[]'))
                VB=$(printf '%s\n' "$(
                    IFS=,
                    printf '%s' "${VB[*]}"
                )")
                VB=${VB//','/' '}
            fi

            printf '%-38s%-22s%-39s%-13s%-33s%-60s\n' "${NAMES[$i]}" "${SN}" "${APIGROUPVERSION}" "${NAMESPACED[$i]}" "${KINDS[$i]}" "[${VB}]"
        done
    fi
done
