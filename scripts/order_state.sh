#!/bin/bash

# 检查参数数量
if [ "$#" -ne 3 ]; then
  echo "Usage: $0 <start> <end> <interval>"
  exit 1
fi

# 获取脚本参数
start=$1
end=$2
interval=$3

# 检查参数是否为数字
if ! [[ "$start" =~ ^[0-9]+$ ]] || ! [[ "$end" =~ ^[0-9]+$ ]] || ! [[ "$interval" =~ ^[0-9]+$ ]]; then
  echo "Error: All parameters must be positive integers."
  exit 1
fi

# 确保 start 小于等于 end
if [ "$start" -gt "$end" ]; then
  echo "Error: start must be less than or equal to end."
  exit 1
fi

# 循环请求并处理结果
current=$start
while [ "$current" -le "$end" ]; do
  # 发起请求并提取 state_root
  hash=$(rooch rpc request --method rooch_getTransactionsByOrder --params "[$current]" | jq -r '.data[0].execution_info.state_root')
  
  # 检查返回值是否为空
  if [ -z "$hash" ]; then
    echo "$current:<no_hash>"
  else
    echo "$current:$hash"
  fi

  # 增加 interval
  current=$((current + interval))
done

# 确保 end 被请求一次（如果不是 interval 的倍数）
if [ $(( (end - start) % interval )) -ne 0 ]; then
  hash=$(rooch rpc request --method rooch_getTransactionsByOrder --params "[$end]" | jq -r '.data[0].execution_info.state_root')
  if [ -z "$hash" ]; then
    echo "$end:<no_hash>"
  else
    echo "$end:$hash"
  fi
fi
