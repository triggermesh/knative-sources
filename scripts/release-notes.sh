#!/usr/bin/env sh

RELEASE=${1:-${GIT_TAG}}
RELEASE=${RELEASE:-${CIRCLE_TAG}}

if [ -z "${RELEASE}" ]; then
	echo "Usage:"
	echo "./scripts/release-notes.sh v0.1.0"
	exit 1
fi

if ! git rev-list ${RELEASE} >/dev/null 2>&1; then
	echo "${RELEASE} does not exist"
	exit
fi

BASE_URL="https://github.com/triggermesh/knative-sources/releases/download/${RELEASE}"
PREV_RELEASE=${PREV_RELEASE:-$(git describe --tags --abbrev=0 ${RELEASE}^ 2>/dev/null)}
PREV_RELEASE=${PREV_RELEASE:-$(git rev-list --max-parents=0 ${RELEASE}^ 2>/dev/null)}
NOTABLE_CHANGES=$(git cat-file -p ${RELEASE} | sed '/-----BEGIN PGP SIGNATURE-----/,//d' | tail -n +6)
CHANGELOG=$(git log --no-merges --pretty=format:'- [%h] %s (%aN)' ${PREV_RELEASE}..${RELEASE})
if [ $? -ne 0 ]; then
	echo "Error creating changelog"
	exit 1
fi

SOURCES=$(sed -n -e "s/^\(SUBDIRS[[:space:]]*?=[[:space:]]*\)\(.*\)$/\2/p" $(dirname ${0})/../Makefile)
COMMANDS=$(sed -n -e "s/^\(COMMANDS[[:space:]]*=[[:space:]]*\)\(.*\)$/\2/p" $(dirname ${0})/inc.Makefile)
PLATFORMS=$(sed -n -e "s/^\(SOURCES[[:space:]]*?=[[:space:]]*\)\(.*\)$/\2/p" $(dirname ${0})/inc.Makefile)

RELEASE_ASSETS_TABLE=$(
  echo -n "| source |"; for command in ${COMMANDS}; do echo -n " ${command} |"; done ; echo
  echo -n "|--|"; for command in ${COMMANDS}; do echo -n "--|"; done ; echo
  for source in ${SOURCES}; do
    echo -n "| $source |"
    for command in ${COMMANDS}; do
      echo -n " ([container](https://gcr.io/triggermesh/${source}-source-${command}:${RELEASE}))"
      for platform in ${PLATFORMS}; do
        echo -n " ([${platform}](${BASE_URL}/${source}-${command}-${platform%/*}-${platform#*/}))"
      done
      echo -n " |"
    done
    echo
  done
)

cat <<EOF
${NOTABLE_CHANGES}

## Installation

Download Knative sources ${RELEASE}

${RELEASE_ASSETS_TABLE}

## Changelog

${CHANGELOG}
EOF
