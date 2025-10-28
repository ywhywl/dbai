#!/bin/bash

# 定义变量
IP_FILE="ip_list.txt"
SOURCE_FILE="/path/to/source/file"
DEST_PATH="/path/to/destination/"

# 检查IP文件是否存在
if [ ! -f "$IP_FILE" ]; then
    echo "错误: IP文件 $IP_FILE 不存在"
    exit 1
fi

# 逐行读取IP文件
while IFS= read -r ip; do
    # 跳过空行和注释行
    if [[ -z "$ip" || "$ip" =~ ^# ]]; then
        continue
    fi
    
    # 提取第三位IP段
    third_octet=$(echo "$ip" | cut -d. -f3)
    
    # 判断是否为90
    if [ "$third_octet" -eq 90 ]; then
        echo "处理IP: $ip (第三位为90)"
        ansible all -m copy -a "src=$SOURCE_FILE dest=$DEST_PATH"
    else
        echo "跳过IP: $ip (第三位为$third_octet)"
    fi
done < "$IP_FILE"
