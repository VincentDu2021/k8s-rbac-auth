############################################################################################################################
# Intel Barcelona datacenter internal use only.
#
# usage: see ./refresh-k8s-token.sh -h
# Before using this script, first verify the "_api_server_addr", "_cluster_name" and "_cluster_ca" are set properly in the
# "#Default values" section. Next, the following values must have been obtained:
# 1. username and password: your linux credential
# 2. oidc-issuer-url: provided to you by RA-2 system admin
# 3. oidc-client-id: provided to you by RA-2 system admin
# 4. oidc-client-secret: provided to you by RA-2 system admin
#
# Disclaimer:
# This script uses above inputs to obtain OIDC id-token and refresh-token from the oidc-provider, then use them to configure
# kubectl. Since OIDC tokens are meant to be short-lived, convenience is hereby provided:
#
# Use '-r' option to have the prompt for your username and password, without being echo-ed back to the terminal.
# Use '-w' option in the first time running this script or when there is update to above values, the provided iput, including
# your password will be written into a "client_file" in  $HOME/.kube/.$username.k8s-oidc-client, then it is base64-encoded with
# permission set to 400. Once existing session expires, the associated tokens will be invalidated, you may execute this script
# without any input, as it reads the needed inputs from the client file instead.
#
# If this is not desired however, i.e. you do not prefer your password being stored this way, you may avoid usig '-w' option,
# instead provide your password in "-p password" everytime you run this script, other iput parameters work in the same manner.
#
# Avoid using this script in any sort of autoamtion, as it may open too many sessions to Keycloak.
###############################################################################################################################

#! /bin/bash

function _cleanup() {
  unset -f _usage _cleanup _update_env _write_client _fetch_token _ write_kubeconfig
  return 0
}

## Clear out nested functions on exit
trap _cleanup INT EXIT RETURN

# Default values
_write_needed=false
_dry_run=false
_version_string="1.0.0"
_api_server_addr="https://10.18.12.10:6443"
_cluster_name="prod"
_cluster_ca="/mnt/weka/weka-csi/mlops/cluster-configs/$_cluster_name/ca.crt"

function _usage() {
  cat <<"EOF"
  ***************************************************************************
  Usage: ./refresh-k8s-token.sh [<option> <parameter>][<option>]...
  Options:
    -u   username of oidc client, expects single arg
    -p   password of oidc client, expects single arg
    -r   prompt for username and password
    -l   oidc issuer url for token fetch and validate, expects single arg
    -i   oidc client id, expects single arg
    -x   oidc client secret, expects single arg
    -w   write into the client file, no arg
    -d   dry-run, writes client file, fetches and prints tokens, no arg
    -v   show version, no arg
    -h   print this message
  ****************************************************************************

  Examples:
  1. For the very first run, writes a client file as $HOME/.kube/.$username.k8s-oidc-client,
     fetches oidc tokens and updates kubeconfig:

     ./refresh-k8s-token.sh \
         -r \
         -l https://example.com/auth/realms/test-realm \
         -i kubernetes \
         -x 398ebeef-feed-1234-4321-feadbeeffead \
         -w
     follow the prompt to input username and password, or,
     ./refresh-k8s-token.sh \
         -u username \
         -p password \
         -l https://example.com/auth/realms/test-realm \
         -i kubernetes \
         -x 398ebeef-feed-1234-4321-feadbeeffead \
         -w

  2. With a client file, run without arguments to update kubeconf with new token:

     ./refresh-k8s-token.sh

  3. Update oidc client ID and Secret With only -i and -x flags and new values provided,
     -w option will update the client file:

     ./refresh-k8s-token.sh \
         -i new-oidc-client-id \
         -x new-oidc-cleint-secret \
         -w
EOF
}

function _take_input() {
  read -s -p "Username: " _username
  echo ""
  read -s -p "Password: " _password
  echo ""
}

# Loads existing client file and updates env with overwrites by user input
function _update_env() {
  # Need to at least have a username, if user does not specify will use login user
  if [[ -z $_username ]]; then
    client_file=$HOME/.kube/.$USER.k8s-oidc-client
  else
    client_file=$HOME/.kube/.$_username.k8s-oidc-client
  fi

  if [[ ! -f $client_file ]]; then
    if [[ $1 == "write" ]]; then
      touch $client_file
    else
      return 0
    fi
  fi

  . <(base64 -d $client_file)
  [[ -z ${_username} ]] || {
    KC_USERNAME=$_username
    _write_needed=true
  }
  [[ -z ${_password} ]] || {
    KC_PASSWORD=$_password
    _write_needed=true
  }
  [[ -z ${_oidc_url} ]] || {
    OIDC_ISSUER_URL=$_oidc_url
    _write_needed=true
  }
  [[ -z ${_oidc_cid} ]] || {
    OIDC_CLIENT_ID=$_oidc_cid
    _write_needed=true
  }
  [[ -z ${_oidc_csc} ]] || {
    OIDC_CLIENT_SC=$_oidc_csc
    _write_needed=true
  }
}

# Writes client file with user input
function _write_client() {
  echo "Writing client ..."
  [[ -d $HOME/.kube ]] || mkdir -p $HOME/.kube
  _update_env "write"

  if [[ $_write_needed ]]; then
    echo "Backup existing $client_file to $client_file.bak"
    mv -f $client_file $client_file.bak

    echo "Writing into $client_file."
    base64 >$client_file <<EOF
# Please treat this file same as to a private key(400) as it contains sensitive info
KC_USERNAME=$KC_USERNAME
KC_PASSWORD='$KC_PASSWORD'
OIDC_ISSUER_URL=$OIDC_ISSUER_URL
OIDC_CLIENT_ID=$OIDC_CLIENT_ID
OIDC_CLIENT_SC=$OIDC_CLIENT_SC
EOF
    chmod 400 $client_file
  else
    echo "No change to be written to client file."
  fi
}

# Login to oidc_token_server to fetch tokens
function _fetch_token() {
  echo "Fetching token:"
  _update_env

  TOKEN=$(curl -s ${OIDC_ISSUER_URL}/protocol/openid-connect/token \
    -d grant_type=password \
    -d response_type=id_token \
    -d scope=openid \
    -d client_id=${OIDC_CLIENT_ID} \
    -d client_secret=${OIDC_CLIENT_SC} \
    -d username=${KC_USERNAME} \
    -d password=${KC_PASSWORD})

  RET=$?
  if [[ "$RET" != "0" ]]; then
    echo "# Error ($RET) ==> ${TOKEN}"
    exit ${RET}
  fi

  ERROR=$(echo ${TOKEN} | jq .error -r)
  if [[ "${ERROR}" != "null" ]]; then
    echo "# Failed ==> ${TOKEN}" >&2
    exit 1
  fi

  ID_TOKEN=$(echo ${TOKEN} | jq .id_token -r)
  REFRESH_TOKEN=$(echo ${TOKEN} | jq .refresh_token -r)
  echo "Successfully fetched tokens."
  $_dry_run && echo -e "id_token=$ID_TOKEN\n----------------\nrefresh_token=$REFRESH_TOKEN"
}

# Write tokens to kubeconfig file
function _write_kubeconfig() {
  echo "Setting up kubectl config:"

  # the cluster name, ip. etc. here are hard-coded, as typical user would only access 1 cluster.
  # If cluster changes, make sure to modify the default values for these entries:
  # 1. $CLUSTER
  # 2. --server
  # 3. cluster CA

  CLUSTER=$_cluster_name
  _kube_cfg="$HOME/.kube/$KC_USERNAME-$CLUSTER.config"
  [[ -f $_kube_cfg ]] || touch $_kube_cfg
  KUBECTL_CONFIG_SET="kubectl config --kubeconfig $_kube_cfg"

  export KUBECONFIG=$_kube_cfg

  $KUBECTL_CONFIG_SET set-cluster \
    $CLUSTER \
    --server=$_api_server_addr

  $KUBECTL_CONFIG_SET set clusters.$CLUSTER.certificate-authority-data \
    $(cat ${_cluster_ca} | base64 -w 0)

  $KUBECTL_CONFIG_SET set-context \
    "$KC_USERNAME-$CLUSTER-context" \
    --cluster=$CLUSTER \
    --user=$KC_USERNAME

  $KUBECTL_CONFIG_SET use-context \
    "$KC_USERNAME-$CLUSTER-context"

  $KUBECTL_CONFIG_SET set-credentials \
    $KC_USERNAME \
    --auth-provider=oidc \
    --auth-provider-arg=idp-issuer-url=$OIDC_ISSUER_URL \
    --auth-provider-arg=client-id=$OIDC_CLIENT_ID \
    --auth-provider-arg=client-secret=$OIDC_CLIENT_SC \
    --auth-provider-arg=refresh-token=$REFRESH_TOKEN \
    --auth-provider-arg=id-token=$ID_TOKEN

  [[ $? == 0 ]] && echo "Successfully set kubeconfig."
}

[[ $# == 0 ]] && echo -e "proceed with existing client"

###############################################
#######  Entry point, getops function  ########
while getopts ':u:p:rl:i:x:wdvh' OPTION; do
  case "$OPTION" in
  u) _username=${OPTARG} ;;
  p) _password=${OPTARG} ;;
  r) _take_input ;;
  l) _oidc_url=${OPTARG} ;;
  i) _oidc_cid=${OPTARG} ;;
  x) _oidc_csc=${OPTARG} ;;
  w) _write_client ;;
  d) _dry_run=true ;;
  v)
    echo -e "Version: $_version_string"
    exit 0
    ;;
  h)
    _usage
    exit 0
    ;;
  ?)
    _usage "<<< invalid option: ${OPTARG}>>>"
    exit 1
    ;;
  esac
done
###############################################

_fetch_token
$_dry_run || _write_kubeconfig

echo "All Done."
