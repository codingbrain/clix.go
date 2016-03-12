OUTDIR=_out
PKGS="clix args exts"

env-setup() {
    mkdir -p $OUTDIR
    set -e
}

for-each-pkg() {
    for pkg in $PKGS; do
        PKG=$pkg "$@"
    done
}

COV_REPORT=$OUTDIR/all.cov
: ${COV_MODE:=set}

cov-run-pkg() {
    go test -v \
        -cover \
        -covermode $COV_MODE \
        -coverprofile=$OUTDIR/$PKG.cov \
        ./$PKG/...
}

cov-collect-pkg() {
    if [ -f $OUTDIR/$PKG.cov ]; then
        grep -E -v '^mode:' $OUTDIR/$PKG.cov >>$COV_REPORT
    fi
}

cov-run() {
    for-each-pkg cov-run-pkg
}

cov-collect() {
    echo mode: $COV_MODE >$COV_REPORT
    for-each-pkg cov-collect-pkg
}

cov-report() {
    go tool cover -$1=$COV_REPORT
}
