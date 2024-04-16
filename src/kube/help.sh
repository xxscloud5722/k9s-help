#!/bin/bash

####################################
# 描述: Kubernetes 快速部署脚本
# 作者: 丁坤
# 邮箱: dingk@billbear.cn
# 时间: 2024-04-01
# 版本: 1.0.1
####################################

GL_KUBE_CONF=
GL_KUBE_NAMESPACE=
KUBE_TYPE_CH_NAME=(无状态 有状态 守护进程集 任务 定时任务 服务 路由 配置项 保密字典)
KUBE_TYPE_EN_NAME=(Deployment StatefulSet DaemonSet Job CronJob Service Ingress Configmap Secret)
KUBE_TYPE_DIRECTORY=(./deployment ./statefulSet ./daemonSet ./job ./cronJob ./service ./ingress ./configmap ./secret)

function apply() {
  kubectl --kubeconfig="$GL_KUBE_CONF" apply -n "$GL_KUBN_NAMESPACE" -f "$1"
}

function print_directory() {
  if [ ! -d "$1" ]; then
    return
  fi
  files=$(ls "$1")
  file_count=$(ls "$1" | wc -l)
  echo "● $2 ${file_count} ($3)"
  for file in $files; do
    echo "  - $file"
  done
}

function convert() {
  # 参数检查
  if [ $# -lt 3 ]; then
    echo "Usage: help --convert <options>"
    echo "Options:"
    echo "  --kubeconfig  集群连接YAML"
    echo "  --namespace   集群命名空间"
    exit 1
  fi

  for arg in "$@"; do
    case $arg in
    --kubeconfig=*)
      kubeconfig="${arg#*=}"
      shift
      ;;
    --namespace=*)
      namespace="${arg#*=}"
      shift
      ;;
    *) ;;
    esac
  done

  if [ -z "$kubeconfig" ]; then
    echo -e "\033[0;31m[系统] 集群配置文件错误\033[0m"
    exit 1
  elif [ -z "$namespace" ]; then
    echo -e "\033[0;31m[系统] 命名空间参数错误\033[0m"
    exit 1
  fi

  if [ ! -e "$kubeconfig" ]; then
    echo -e "\033[0;31m[系统] $kubeconfig 配置文件不存在\033[0m"
    exit 1
  fi

  echo -e "\033[1;34m[系统] 迁移集群配置文件: $kubeconfig\033[0m"
  echo -e "\033[1;34m[系统] 迁移集群命名空间: $namespace\033[0m"

  # 扫描配置目录
  echo "[系统] 扫描YAML文件 ..."
  for ((i = 0; i < 9; i++)); do
    ch_name="${KUBE_TYPE_CH_NAME[${i}]}"
    en_name="${KUBE_TYPE_EN_NAME[${i}]}"
    directory="${KUBE_TYPE_DIRECTORY[${i}]}"
    print_directory "$directory" "$ch_name" "$en_name"
  done

  read -p "是否确认执行? (输入 'yes' 继续): " answer
  if [ "$answer" != "yes" ]; then
    echo -e "\033[0;31m[系统] 用户终止执行\033[0m"
    exit 1
  fi

  # 检查集群是否存在命名空间
  if kubectl get namespace --kubeconfig="$kubeconfig" | grep -q "^${namespace}"; then
    echo -e "\033[0;33m[系统] 集群: $kubeconfig, 已存在命名空间: ${namespace}\033[0m"
    read -p "命名空间: ${namespace}已存在, 是否继续执行? (输入 'yes' 继续): " answer
    if [ "$answer" != "yes" ]; then
      echo -e "\033[0;31m[系统] 用户终止执行\033[0m"
      exit 1
    fi
  else
    # 创建命名空间
    echo "[系统] 创建命名空间: ${namespace}"
    kubectl create namespace "$namespace" --kubeconfig="$kubeconfig"
  fi

  # 执行YAML文件
  for ((i = 0; i < 9; i++)); do
    directory="${KUBE_TYPE_DIRECTORY[${i}]}"
    en_name="${KUBE_TYPE_EN_NAME[${i}]}"
    if [ ! -d "$directory" ]; then
      continue
    fi
    files=$(ls "$directory")
    for file in $files; do
      yaml="$directory/$file"
      echo "[系统] YAML更新: $directory/$file"
      if [ "$en_name" == "Service" ]; then
        service_yaml="/tmp/$(uuidgen).yaml"
        yq eval 'del(.spec.ports[] | select(.nodePort != null) | .nodePort)' "$yaml" >"$service_yaml"
        kubectl --kubeconfig="$kubeconfig" apply -n "$namespace" -f "$service_yaml"
      else
        kubectl --kubeconfig="$kubeconfig" apply -n "$namespace" -f "$yaml"
      fi
    done
  done
}

# 脚本入口函数
if [ $# -lt 1 ]; then
  echo "Kubernetes 快速部署脚本 版本: 1.0.1"
  echo "Usage: help <options>"
  echo "Options:"
  echo "  --convert    执行命名空间迁移集群"
  echo "  [YAML File]  执行当前集群YAML"
  exit 1
fi

option=$1

case $option in
"--convert")
  echo "[系统] 执行命名空间迁移集群 ..."
  echo -e "\e[33m[系统] 不会迁移命名空间下PVC 和 PV, 请人工处理\e[0m"
  convert "$@"
  ;;
*)
  echo "[系统] 执行当前集群命名空间YAML ..."
  apply "$1"
  ;;
esac
