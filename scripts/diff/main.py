from argparse import ArgumentParser
from typing import Tuple
import difflib

def parse_filename_from_args() -> Tuple[str, str]:
    arg_parser = ArgumentParser()
    arg_parser.add_argument("--client", type=str, help="client query results file path", required=True, metavar="client")
    arg_parser.add_argument("--pandas", type=str, help="pandas query results file path", required=True, metavar="pandas")    
    args = arg_parser.parse_args()
    return args.client, args.pandas

def parse_querys(lines: list[str]) -> dict[str, str]:
    results = dict()    
    assert lines[0].rstrip() == "==========="
    i = 0
    while i < len(lines):
        query = lines[i+1].rstrip().removesuffix(":")
        i+=3
        s = list()
        while i < len(lines)  and lines[i] != "===========\n":
            s.append(lines[i])
            i+=1
        results[query] = s

    return results
        
def main():
    client_filename, pandas_filename = parse_filename_from_args()
    with open(client_filename, "r") as f:
        client_lines = f.readlines()

    client_results = parse_querys(client_lines)
    
    with open(pandas_filename, "r") as f:
        pandas_lines = f.readlines()

    pandas_results = parse_querys(pandas_lines)

    # Expensive but worth it
    assert set(client_results.keys()) == set(pandas_results.keys())

    for k in client_results.keys():
        diff_result  = difflib.ndiff(client_results[k], pandas_results[k])
        print(f"{k}\n {''.join(diff_result)}")
        
if "__main__" == __name__:
    main()
