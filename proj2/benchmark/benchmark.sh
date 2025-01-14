#!/bin/bash
#
#SBATCH --mail-user=nichada@cs.uchicago.edu
#SBATCH --mail-type=ALL
#SBATCH --job-name=proj2_grade
#SBATCH --output=./slurm/out/%j.%N.stdout
#SBATCH --error=./slurm/out/%j.%N.stderr
#SBATCH --chdir=/home/nichada/MPCSParallelpgm/project-2-pannich/proj2/benchmark
#SBATCH --partition=debug
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=16
#SBATCH --mem-per-cpu=900
#SBATCH --exclusive
#SBATCH --time=100:00

module load golang/1.19
mkdir -p ./slurm/out

# Array of thread numbers to test
# sizes=("large" "xlarge")
sizes=("xsmall" "small" "medium" "large" "xlarge")
threads=(1 2 4 6 8 12)
partypes=("s" "p")

# Directory where you want to save the output times
output_dir="./times"
mkdir -p "$output_dir/mintimes"

# Compile editor
go build -o ../benchmark/benchmark ../benchmark/benchmark.go

# Function to run the editor and record the minimum time out of five attempts
run_and_record_min_time() {
    local par=$1
    local size=$2
    local thread_count=$3
    local times_file="$output_dir/mintimes/times_${par}_${size}_${thread_count}.txt"
    > "$times_file"  # Clear or create the file to store times

    # Repeat the timing five times and save each real time
    for i in {1..5}; do
        # 1>> stdout to log.txt
        # 2>&1 redirect stdout and strerr to next command
        # awk send first line (timer by go) to $time_file (use this)
        # there will be unix time output in stdout (not used)
        if [[ $par == "s" ]]; then
          { time ../benchmark/benchmark s "$size" 1>>"$times_file";}
        else
          { time ../benchmark/benchmark p "$size" "$thread_count" 1>>"$times_file"; }
        fi
    done

    # Minimum time
    min_time=$(sort -h "$times_file" | head -n1)

    # Write the minimum time to a summary file `times_summary.txt`
    echo "Min Time: $min_time" >> "$output_dir/times_summary.txt"
}

# Loop over the thread numbers and record the time for each
echo " Record Test from go timer " > "$output_dir/times_summary.txt"

for par in "${partypes[@]}"; do
  for size in "${sizes[@]}"; do
      if [[ $par == "s" ]]; then
        echo "Sequential size: $size" >> "$output_dir/times_summary.txt"
        run_and_record_min_time "$par" "$size" 0
      else
        for t in "${threads[@]}"; do
          echo "Parallel size: $size threads: $t" >> "$output_dir/times_summary.txt"
          run_and_record_min_time "$par" "$size" "$t"
        done
      fi
  done
done
