#!/bin/bash
#
#SBATCH --mail-user=CNETID@cs.uchicago.edu
#SBATCH --mail-type=ALL
#SBATCH --job-name=proj2_benchmark
#SBATCH --output=./slurm/out/%j.%N.stdout
#SBATCH --error=./slurm/out/%j.%N.stderr
#SBATCH --chdir=ABSOLUTE_PATH_TO_PROJ1_GRADER_DIRECTORY
#SBATCH --partition=debug
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=16
#SBATCH --mem-per-cpu=900
#SBATCH --exclusive
#SBATCH --time=5:00


module load golang/1.19
go test proj2/twitter -v -count=1
