# SPDX-License-Identifier: Apache-2.0 OR MIT
#
# Rendered by `make demo` (vhs, https://github.com/charmbracelet/vhs) from
# the repository root, which sets PAIN001_SERVER and DEMO_PORT. Hidden
# from the recording: copy the tracked files into a throwaway directory,
# start `pain001 serve` on 127.0.0.1 and wait for /api/v1/health. On
# screen: the typed live tests call Health, Validate, Generate and Check,
# then examples/basic generates pain.001.001.09 from the conformance
# kit's fixture and the file's header is shown. Only loopback is used.
Output .github/demo.gif

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 640
Set Padding 20
Set Theme "Catppuccin Mocha"

Hide
Type `D="$(mktemp -d /tmp/pain001-sdk-go-demo.XXXX)" && git ls-files -z | tar --null -T - -cf - | tar -xf - -C "$D" && cd "$D"`
Enter
Type `("$PAIN001_SERVER" serve --host 127.0.0.1 --port "$DEMO_PORT" > server.log 2>&1 & echo $! > server.pid)`
Enter
Type `export PAIN001_BASE_URL="http://127.0.0.1:$DEMO_PORT"; for _ in $(seq 60); do curl -fsS "$PAIN001_BASE_URL/api/v1/health" > /dev/null 2>&1 && break; sleep 1; done`
Enter
Type `go build ./... && go test -count=1 -run '^$' . > /dev/null && export PS1='$ ' && clear`
Enter
Wait@120s /^\$$/
Show

Type "go test -count=1 -v -run 'TestLive(Health|Validate|Generate|Check)$' ."
Sleep 500ms
Enter
Wait@60s /^\$$/
Sleep 3s

Type "go run ./examples/basic -out payments.xml"
Sleep 500ms
Enter
Wait@60s /^\$$/
Sleep 3s

Type "head -n 12 payments.xml"
Sleep 500ms
Enter
Wait /^\$$/
Sleep 5s

Hide
Type `kill "$(cat server.pid)"; cd / && rm -rf "$D"`
Enter
Sleep 1s
