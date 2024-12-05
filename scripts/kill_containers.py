import subprocess
import random
import time
import argparse
from datetime import datetime


parser = argparse.ArgumentParser(prog="kill_containers", description="randomly kill containers in random intervals")
parser.add_argument("--min", default=5)
parser.add_argument("--max", default=10)
args = parser.parse_args()
# Config parameters
min_time = int(args.min)
max_time = int(args.max)

non_killable_containers = ["healthcheck_1", "boundary", "rabbitmq"]

output = subprocess.run(
    ["docker", "container", "ls", "--format", "{{.Names}}"], capture_output=True
)
containers = output.stdout.decode().splitlines()
killable_containers = [
    container for container in containers if container not in non_killable_containers
]
gen_random_time = lambda: random.randint(min_time, max_time)

waiting_time = gen_random_time()
while True:
    container_to_kill = random.choice(killable_containers)
    subprocess.run(
        ["docker", "container", "kill", container_to_kill], stdout=subprocess.DEVNULL
    )
    print(f"{datetime.now():%Y-%m-%d %H:%M:%S} - killed container: {container_to_kill}")
    time.sleep(waiting_time)
    waiting_time = gen_random_time()
