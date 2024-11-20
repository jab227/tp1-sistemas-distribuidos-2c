from argparse import ArgumentParser
from typing import Tuple
import difflib


def parse_filename_from_args() -> Tuple[str, str]:
    arg_parser = ArgumentParser()
    arg_parser.add_argument(
        "--first",
        type=str,
        help="first query results path",
        required=True,
        metavar="first",
    )
    arg_parser.add_argument(
        "--second",
        type=str,
        help="second query results path",
        required=True,
        metavar="second",
    )
    args = arg_parser.parse_args()
    return args.first, args.second


def parse_querys(lines: list[str]) -> dict[str, str]:
    results = dict()
    assert lines[0].rstrip() == "==========="
    i = 0
    while i < len(lines):
        query = lines[i + 1].rstrip().removesuffix(":")
        i += 3
        s = list()
        while i < len(lines) and lines[i] != "===========\n":
            s.append(lines[i])
            i += 1
        results[query] = s

    return results


def main():
    first_filename, second_filename = parse_filename_from_args()
    with open(first_filename, "r") as f:
        first_lines = f.readlines()

    first_results = parse_querys(first_lines)

    with open(second_filename, "r") as f:
        second_lines = f.readlines()

    second_results = parse_querys(second_lines)

    # Expensive but worth it
    assert set(first_results.keys()) == set(second_results.keys())

    for k in first_results.keys():
        diff_result = difflib.ndiff(first_results[k], second_results[k])
        print(f"{k}\n{''.join(diff_result)}")


if "__main__" == __name__:
    main()
