angular.module('gsApplication', [])
    .controller('MainCtrl', [ '$scope', '$http', '$sce', '$timeout', function($scope, $http, $sce, $timeout) {
        window.debug = { scope: $scope };
        var urlPrefix = "http://localhost:5335";
        var POLL_PERIOD = 250;

        $scope.highlightCode = function(str) {
            var args;
            var tokens = str.split(" ", 2);
            var op = tokens[0];
            if (typeof tokens[1] === "undefined") {
                args = "";
            } else {
                args = tokens[1];
                args = args.replace(/(\(|\))/g, '<span class="hl hl-prn">$1</span>')
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
            // TODO
            return false;
        };

        $scope.processDisassembly = function() {
            $scope.disasm.entries.map(function(elem) {
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
        };

        $scope.loadDisassembly = function(addr) {
            if (typeof addr === 'number') {
                addr = addr.toString(16);
            }
            $http.get(urlPrefix + "/dump/disasm?start=" + addr).then(
                function(data) {
                    $scope.disasm = data.data;
                    $scope.processDisassembly();
                }
            )
        };

        $scope.loadBreakpoints = function() {
            $http.get(urlPrefix + "/breakpoints").then(
                function(data) {
                    $scope.breakpoints = {};
                    data.data.forEach(function(addr) {
                        $scope.breakpoints[addr] = true;
                    });
                }
            )
        };

        $scope.pollRegisters = function() {
            var loadRegisters = function() {
                $http.get(urlPrefix + "/dump/registers").then(
                    function(data) {
                        $scope.r = data.data;
                        if ($scope.r.mode === "run") {
                            $scope.triggerText = "stop";
                        } else {
                            $scope.triggerText = "resume";
                        }
                        $scope.regTimeout = $timeout(loadRegisters, POLL_PERIOD);
                    },
                    function(err) {
//                        $scope.regTimeout = $timeout(loadRegisters, POLL_PERIOD);
                    }
                );
            };
            if ($scope.regTimeout) {
                $timeout.cancel($scope.regTimeout);
            }
            $scope.regTimeout = $timeout(loadRegisters, POLL_PERIOD);
        };

        $scope.stop = function() {
            $http.post(urlPrefix + "/control/stop", {})
                .then(function() {
                    $scope.pollRegisters();
                    $scope.loadDisassembly($scope.r.PC);
                });
        };

        $scope.step = function() {
            $http.post(urlPrefix + "/control/step", {})
                .then(function() {
                    $scope.pollRegisters();
                    $scope.loadDisassembly($scope.r.PC); // TODO: change to shouldAddDisassembly()
                });
        };

        $scope.resume = function() {
            $http.post(urlPrefix + "/control/resume", {})
                .then(function() {
                    // resumed
                });
        };

        $scope.enableBreakpoints = function() {
            $http.post(urlPrefix + "/control/enable_bp", {})
                .then(function() {
                    $scope.pollRegisters();
                });
        };

        $scope.disableBreakpoints = function() {
            $http.post(urlPrefix + "/control/disable_bp", {})
                .then(function() {
                    $scope.pollRegisters();
                });
        };

        $scope.addBreakpoint = function(addr) {
            $http.post(urlPrefix + "/breakpoints/" + addr, {}).then(
                function(data) {
                    $scope.loadBreakpoints();
                }
            )
        };

        $scope.removeBreakpoint = function(addr) {
            if (!(addr in $scope.breakpoints)) {
                return;
            }
            $http.delete(urlPrefix + "/breakpoints/" + addr, {}).then(
                function(data) {
                    $scope.loadBreakpoints();
                }
            )
        };



        $scope.disasm = {"entries":[{"addr":"0000","code":"DI","bytes":["F3"]},{"addr":"0001","code":"XOR A","bytes":["AF"]},{"addr":"0002","code":"LD DE,0xFFFFh","bytes":["11","FF","FF"]},{"addr":"0005","code":"JP (0x11CBh)","bytes":["C3","CB","11"]},{"addr":"0008","code":"LD HL,(0x5C5Dh)","bytes":["2A","5D","5C"]},{"addr":"000B","code":"LD (0x5C5Fh),HL","bytes":["22","5F","5C"]},{"addr":"000E","code":"JR (PC+67)","bytes":["18","43"]},{"addr":"0010","code":"JP (0x15F2h)","bytes":["C3","F2","15"]},{"addr":"0013","code":"RST 38H","bytes":["FF"]},{"addr":"0014","code":"RST 38H","bytes":["FF"]},{"addr":"0015","code":"RST 38H","bytes":["FF"]},{"addr":"0016","code":"RST 38H","bytes":["FF"]},{"addr":"0017","code":"RST 38H","bytes":["FF"]},{"addr":"0018","code":"LD HL,(0x5C5Dh)","bytes":["2A","5D","5C"]},{"addr":"001B","code":"LD A,(HL)","bytes":["7E"]},{"addr":"001C","code":"CALL (0x007Dh)","bytes":["CD","7D","00"]},{"addr":"001F","code":"RET NC","bytes":["D0"]},{"addr":"0020","code":"CALL (0x0074h)","bytes":["CD","74","00"]},{"addr":"0023","code":"JR (PC+-9)","bytes":["18","F7"]},{"addr":"0025","code":"RST 38H","bytes":["FF"]},{"addr":"0026","code":"RST 38H","bytes":["FF"]},{"addr":"0027","code":"RST 38H","bytes":["FF"]},{"addr":"0028","code":"JP (0x335Bh)","bytes":["C3","5B","33"]},{"addr":"002B","code":"RST 38H","bytes":["FF"]},{"addr":"002C","code":"RST 38H","bytes":["FF"]},{"addr":"002D","code":"RST 38H","bytes":["FF"]},{"addr":"002E","code":"RST 38H","bytes":["FF"]},{"addr":"002F","code":"RST 38H","bytes":["FF"]},{"addr":"0030","code":"PUSH BC","bytes":["C5"]},{"addr":"0031","code":"LD HL,(0x5C61h)","bytes":["2A","61","5C"]},{"addr":"0034","code":"PUSH HL","bytes":["E5"]},{"addr":"0035","code":"JP (0x169Eh)","bytes":["C3","9E","16"]},{"addr":"0038","code":"PUSH AF","bytes":["F5"]},{"addr":"0039","code":"PUSH HL","bytes":["E5"]},{"addr":"003A","code":"LD HL,(0x5C78h)","bytes":["2A","78","5C"]},{"addr":"003D","code":"INC HL","bytes":["23"]},{"addr":"003E","code":"LD (0x5C78h),HL","bytes":["22","78","5C"]},{"addr":"0041","code":"LD A,H","bytes":["7C"]},{"addr":"0042","code":"OR L","bytes":["B5"]},{"addr":"0043","code":"JR NZ,(PC+3)","bytes":["20","03"]},{"addr":"0045","code":"INC (IY+0x40h)","bytes":["FD","34","40"]},{"addr":"0048","code":"PUSH BC","bytes":["C5"]},{"addr":"0049","code":"PUSH DE","bytes":["D5"]},{"addr":"004A","code":"CALL (0x386Eh)","bytes":["CD","6E","38"]},{"addr":"004D","code":"POP DE","bytes":["D1"]},{"addr":"004E","code":"POP BC","bytes":["C1"]},{"addr":"004F","code":"POP HL","bytes":["E1"]},{"addr":"0050","code":"POP AF","bytes":["F1"]},{"addr":"0051","code":"EI","bytes":["FB"]},{"addr":"0052","code":"RET","bytes":["C9"]}],"stack":["0015","E100","3B15","7F0F","5410","B4FF","0012","003E","423C","7E42","","","","","","","","","",""]};
        $scope.r = {"PC":"0017","SP":"FF48","AF":"005C","BC":"0000","DE":"5CB9","HL":"10A8","IX":"0000","IY":"5C3A","AFx":"0044","BCx":"174B","DEx":"0006","HLx":"107F","R":"59","I":"3F","IFF1":true,"IFF2":true,"IM":1,"breakpoints_enabled":false,"cpu_state":"stopped","mode":"run","stack":[{"addr":"FF48","data":"15FE"},{"addr":"FF4A","data":"0000"},{"addr":"FF4C","data":"15E1"},{"addr":"FF4E","data":"0F3B"},{"addr":"FF50","data":"107F"},{"addr":"FF52","data":"FF54"},{"addr":"FF54","data":"12B4"},{"addr":"FF56","data":"3E00"},{"addr":"FF58","data":"3C00"},{"addr":"FF5A","data":"4242"}]};
        $scope.triggerText = "stop";
        $scope.processDisassembly();
        $scope.loadDisassembly(0);
        $scope.pollRegisters();

    }]);