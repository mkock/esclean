# ESclean

ESclean is a helpful tool that allows you to track down and remove unused exports in EcmaScript-compatible
JavaScript projects.

The goal is to support both JavaScript and TypeScript source code analysis.

## Usage

Run: `esclean path/to/indexFile.ts|js`

It will stream a list of unused exports to stdout. _Please double check in your IDE that they aren't used before
removing them._

## How it works

Given an index file, the algorithm traverses the entire source code hierarchy while ignoring third-party packages,
following import paths as far as possible. Each import is checked against a matching export statement from the source
and, if matched, a reference counter is incremented. Finally, a report is generated containing all unmatched exports.

In the early stages of this project, some false positives must be expected. But once a list of candidates exist, it
should be faily easy to double check them in your IDE before removing any unused exports.

## Current limitations

- Does not handle multiline imports / exports
- Files imported by alias (namespaced) are assumed to have been used
- Does not work with CommonJS
- Has no notion of "TypeScript" or "JavaScript" mode, which could give unpredictable results
- Does not check if imports are actually used
- Generated file hashes were introduced to improve lookup speeds, but are currently unused

## Disclaimer

The author does not take responsibility for broken code that results from an analysis made by ESclean.
A fair amount of thoughtfulness must be expected.

## License

This software is distributed under the _MIT_ license. See `LICENSE` for details.
