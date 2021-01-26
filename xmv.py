#! /usr/bin/env python3

from json import load, dumps
from os import environ, path
from sys import argv, exit
import subprocess


config_file = path.join(environ["HOME"], ".config/xmv")

config = {
    "patterns": {}
}


def run_zsh(command):
    return subprocess.run(["/usr/bin/zsh", "-ic", command], capture_output=True, text=True)

def run_all_patterns(config):
    successful_patterns = []

    for old_pattern, new_pattern in config["patterns"].items():
        result = run_zsh(f"zmv -n '{old_pattern}' '{new_pattern}'")

        if result.returncode != 0:
            continue

        samples = result.stdout.split("\n")
        successful_patterns.append((
            old_pattern,
            new_pattern,
            samples,
        ))

    with open(config_file, "w") as file:
        file.write(dumps(config))
    file.close()

    for index, pattern in enumerate(successful_patterns):
        old_pattern, new_pattern, samples = pattern
        print(f"{index + 1}. [{old_pattern} -> {new_pattern}] {samples[0]}")

    print("q. quit\n")

    return successful_patterns

def choose_pattern(successful_patterns):
    chosen_pattern_index = input("choose option: ")

    if chosen_pattern_index.lower() in ["q", ""]:
        exit(0)

    if not chosen_pattern_index.isnumeric():
        exit(1)

    chosen_pattern_index = int(chosen_pattern_index)

    if chosen_pattern_index > len(successful_patterns):
        exit(1)

    return successful_patterns[chosen_pattern_index - 1]

def run_one_pattern(pattern):
    old_pattern, new_pattern, samples = pattern

    print("\n".join(samples))

    proceed = input("proceed? y/n ")

    if proceed.lower() != "y":
        exit(0)

    result = run_zsh(f"zmv '{old_pattern}' '{new_pattern}'")

    print(result.stderr)
    print(result.stdout)

    exit(result.returncode)

def main():
    try:
        with open(config_file) as file:
            config = load(file)
        file.close()
    except FileNotFoundError:
        with open(config_file, "w") as file:
            file.write(dumps(config))
        file.close()

    if len(argv) == 3:
        old_pattern = argv[1]
        new_pattern = argv[2]

        config["patterns"][old_pattern] = new_pattern

    successful_patterns = run_all_patterns(config)
    pattern = choose_pattern(successful_patterns)
    run_one_pattern(pattern)

if __name__ == "__main__":
    main()
