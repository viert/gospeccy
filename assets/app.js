angular.module('gsApplication', [])
    .controller('MainCtrl', [ '$scope', '$http', '$sce', '$timeout', function($scope, $http, $sce, $timeout) {
        window.debug = { scope: $scope };

        var POLL_PERIOD = 250;
        $scope.disasm = {};
        $scope.r = {};
        $scope.disasmBounds = {
            min: null,
            max: null
        };

        $scope.highlightCode = function(str) {
            var args;
            var tokens = str.split(" ", 2);
            var op = tokens[0];
            if (typeof tokens[1] === "undefined") {
                args = "";
            } else {
                args = tokens[1];
                args = args.replace(/(\(|\))/g, '<span class="hl hl-prn">$1</span>');
                args = args.replace(/(0x[0-9a-fA-F]{2,4}h)/g, '<span class="hl hl-hex">$1</span>');
                args = args.replace(/\((BC|DE|HL|IX|IY|PC)/g, '(<span class="hl hl-reg">$1</span>');
                args = args.replace('+-', '-');
                if (op === "RET" || op === "JP" || op === "JR") {
                    args = args.replace(/^(NZ|NC|PO|PE|C|P|Z|M)(,?)/, '<span class="hl hl-flag">$1</span>$2')
                } else {
                    args = args.replace(/^(AF'|AF|BC|DE|HL|IX|IY|SP|A|B|C|D|E|H|L)/g, '<span class="hl hl-reg">$1</span>');
                    args = args.replace(/,(AF'|AF|BC|DE|HL|IX|IY|SP|A|B|C|D|E|H|L)/g, ',<span class="hl hl-reg">$1</span>');
                }
            }
            str = "<span class=\"hl hl-cmd\">" + tokens[0] + "</span> " + args;
            str = $sce.trustAsHtml(str);
            return str;
        };

        $scope.isBreakpoint = function(addr) {
            return addr in $scope.breakpoints;
        };

        $scope.shouldAddDisassembly = function() {
            var pc = parseInt($scope.r.PC, 16);
            console.log(pc, $scope.disasmBounds);
            if (pc > $scope.disasmBounds.max) {
                if (pc - $scope.disasmBounds.max < 3) {
                    $scope.loadDisassembly(pc, true);
                } else {
                    $scope.loadDisassembly(pc);
                }
                return;
            }
            if (pc < $scope.disasmBounds.min) {
                $scope.loadDisassembly(pc);
                return;
            }
        };

        $scope.setDisassemblyBounds = function() {
            var firstEntry = $scope.disasm.entries[0];
            var lastEntry = $scope.disasm.entries[$scope.disasm.entries.length-1];
            $scope.disasmBounds.min = parseInt(firstEntry.addr, 16);
            $scope.disasmBounds.max = parseInt(lastEntry.addr, 16);
        };

        $scope.processDisassembly = function(data) {
            data.entries.map(function(elem) {
                elem.code = $scope.highlightCode(elem.code);
                elem.chars = [];
                var char, code;
                elem.bytes.forEach(function(byte) {
                    code = parseInt(byte, 16);
                    if (code < 32) {
                        char = "&middot;";
                    } else {
                        char = String.fromCharCode(code);
                    }
                    elem.chars.push($sce.trustAsHtml(char));
                });
                return elem;
            });
            return data;
        };

        $scope.loadDisassembly = function(addr, append) {
            if (typeof addr === 'number') {
                addr = addr.toString(16);
            }
            $http.get("/dump/disasm?start=" + addr).then(
                function(data) {
                    var incoming = data.data;
                    incoming = $scope.processDisassembly(incoming);
                    if (append) {
                        $scope.disasm.entries = $scope.disasm.entries.concat(incoming.entries);
                    } else {
                        $scope.disasm = incoming;
                    }
                    $scope.setDisassemblyBounds();
                }
            )
        };

        $scope.loadBreakpoints = function() {
            $http.get("/breakpoints").then(
                function(data) {
                    $scope.breakpoints = {};
                    data.data.forEach(function(addr) {
                        $scope.breakpoints[addr] = true;
                    });
                }
            )
        };

        $scope.pollRegisters = function(onLoad) {
            var loadRegisters = function(callback) {
                $http.get("/dump/registers").then(
                    function(data) {
                        $scope.r = data.data;
                        if ($scope.r.mode === "run") {
                            $scope.triggerText = "stop";
                            POLL_PERIOD = 250;
                        } else {
                            $scope.triggerText = "resume";
                            POLL_PERIOD = 10000;
                        }
                        if (typeof callback === "function") {
                            callback();
                        }
                        $scope.regTimeout = $timeout(loadRegisters, POLL_PERIOD);
                    },
                    function(err) {
                        console.log(err);
                    }
                );
            };
            if ($scope.regTimeout) {
                $timeout.cancel($scope.regTimeout);
            }
            loadRegisters(onLoad);
        };

        $scope.triggerRunmode = function() {
            if ($scope.triggerText === "stop")
                $scope.stop();
            else
                $scope.resume();
        };

        $scope.stop = function() {
            $http.post("/control/stop", {})
                .then(function() {
                    $scope.pollRegisters(function() {
                        $scope.loadDisassembly($scope.r.PC);
                    });
                });
        };

        $scope.reset = function() {
            $http.post("/control/reset", {})
                .then(function() {
                    $scope.pollRegisters(function() {
                        $scope.loadDisassembly($scope.r.PC);
                    });
                });
        };

        $scope.step = function() {
            $http.post("/control/step", {})
                .then(function() {
                    $scope.pollRegisters(function() {
                        $scope.shouldAddDisassembly();
                    });
                });
        };

        $scope.resume = function() {
            $http.post("/control/resume", {})
                .then(function() {
                    $scope.pollRegisters()
                });
        };

        $scope.enableBreakpoints = function() {
            $http.post("/control/enable_bp", {})
                .then(function() {
                    $scope.pollRegisters();
                });
        };

        $scope.disableBreakpoints = function() {
            $http.post("/control/disable_bp", {})
                .then(function() {
                    $scope.pollRegisters();
                });
        };

        $scope.addBreakpoint = function(addr) {
            $http.post("/breakpoints/" + addr, {}).then(
                function(data) {
                    $scope.loadBreakpoints();
                }
            )
        };

        $scope.triggerBreakpoint = function(addr) {
            if (addr in $scope.breakpoints) {
                $scope.removeBreakpoint(addr);
            } else {
                $scope.addBreakpoint(addr);
            }
        };

        $scope.removeBreakpoint = function(addr) {
            if (!(addr in $scope.breakpoints)) {
                return;
            }
            $http.delete("/breakpoints/" + addr, {}).then(
                function(data) {
                    $scope.loadBreakpoints();
                }
            )
        };

        $scope.triggerText = "stop";
        $scope.breakpoints = {};
        $scope.loadDisassembly(0);
        $scope.loadBreakpoints();
        $scope.pollRegisters(
            function() {
                $scope.loadDisassembly($scope.r.PC);
            }
        );

    }]);