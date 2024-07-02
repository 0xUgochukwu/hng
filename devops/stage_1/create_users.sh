#!/bin/bash

declare -a users
declare -a groups

function read_input() {
  local file="$1"

  while IFS= read -r line; do
    user=$(echo "$line" | cut -d';' -f1)
    groups_data=$(echo "$line" | cut -d';' -f2 | tr -d '[:space:]')
    users+=("$user")
    groups+=("$groups_data")
  done < "$file"
}


if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <input_file>"
  exit 1
fi

input_file="$1"
echo "Reading input file: $input_file"
read_input "$input_file"

log_file="/var/log/user_management.log"
password_file="/var/secure/user_passwords.txt"

# Create files if they doesn't exist
if [ ! -f "$log_file" ]; then
  touch "$log_file"
fi

if [ ! -f "$password_file" ]; then
  touch "$password_file"
fi

for (( i = 0; i < ${#users[@]}; i++ )); do
  user="${users[$i]}"
  user_groups="${groups[$i]}"

  echo "Creating User: $user"
  
  if id "$user" &>/dev/null; then
    echo "User $user already exists, Skipped" | tee -a "$log_file"
  else
    # Create user
    useradd -m -s /bin/bash "$user"
    if [[ $? -ne 0 ]]; then
      echo "Failed to create user $user" | tee -a "$log_file"
      exit 1
    fi
    echo "User $user created" | tee -a "$log_file"

    # Set password
    password=$(openssl rand -base64 50 | tr -dc 'A-Za-z0-9!?%=' | head -c 10)
    echo "$user:$password" | chpasswd
    if [[ $? -ne 0 ]]; then
      echo "Failed to set password for $user" | tee -a "$log_file"
      exit 1
    fi

    echo "Password for $user set" | tee -a "$log_file"
    echo "$user:$password" >> "$password_file"

    # Create personal group
    if grep -q "^$user:" /etc/group; then
      echo "Personal group $user already exists" | tee -a "$log_file"
    else
      echo "Personal group $user does not exist, creating $user" | tee -a "$log_file"
      groupadd "$user"
      if [[ $? -ne 0 ]]; then
        echo "Failed to create personal group $user" | tee -a "$log_file"
        exit 1
      fi
    fi
      echo "Personal group $user created for $user" | tee -a "$log_file"

    # Add user to personal group
    usermod -aG "$user" "$user"
    if [[ $? -ne 0 ]]; then
      echo "Failed to add $user to $user group" | tee -a "$log_file"
      exit 1
    fi
    echo "Added $user to $user group" | tee -a "$log_file"


    # Add user to other groups
    for group in $(echo "$user_groups" | tr ',' '\n'); do
      if grep -q "^$group:" /etc/group; then
        echo "Group $group already exists" | tee -a "$log_file"
      else
        echo "Group $group does not exist, creating $group" | tee -a "$log_file"
        groupadd "$group"
      
        if [[ $? -ne 0 ]]; then
          echo "Failed to create group $group" | tee -a "$log_file"
          exit 1
        fi
      fi

    usermod -aG "$group" "$user"
    if [[ $? -ne 0 ]]; then
      echo "Failed to add $user to $group group" | tee -a "$log_file"
      exit 1
    fi
    echo "Added $user to $group group" | tee -a "$log_file"
    done
  fi
done



echo "All users created and added successfully!"
echo "Cheers 🍻"
