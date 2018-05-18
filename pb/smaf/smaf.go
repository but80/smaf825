package smaf

// Normalize は、音色データから異常な値を排除し、正常化します。
// 異常が検出された音色の一覧を返します。
func (lib *VM5VoiceLib) Normalize() []*VM35VoicePC {
	if lib.Programs == nil {
		lib.Programs = []*VM35VoicePC{}
	}
	result := []*VM35VoicePC{}
	for i, pc := range lib.Programs {
		if pc == nil {
			pc = &VM35VoicePC{}
			lib.Programs[i] = pc
		}
		if !pc.Normalize() {
			result = append(result, pc)
		}
	}
	return result
}

func normalizeUint32(ok *bool, target *uint32, min, max uint32) {
	if *target < min {
		*target = min
		*ok = false
	}
	if max < *target {
		*target = max
		*ok = false
	}
}

func normalizeString(ok *bool, target *string, def string) {
	if *target == "" {
		*target = def
		*ok = false
	}
}

// Normalize は、音色データから異常な値を排除し、正常化します。
// 元から正常な音色だったときは true を返します。
func (pc *VM35VoicePC) Normalize() bool {
	ok := true

	if pc.Version < VM35FMVoiceVersion_VM35FMVoiceVersion_MIN || VM35FMVoiceVersion_VM35FMVoiceVersion_MAX < pc.Version {
		pc.Version = VM35FMVoiceVersion_VM5
		ok = false
	}
	normalizeString(&ok, &pc.Name, "(undefined)")
	normalizeUint32(&ok, &pc.BankMsb, 0, 127)
	normalizeUint32(&ok, &pc.BankLsb, 0, 127)
	normalizeUint32(&ok, &pc.Pc, 0, 127)
	normalizeUint32(&ok, &pc.DrumNote, 0, 127)
	if pc.VoiceType < VoiceType_VoiceType_MIN || VoiceType_VoiceType_MAX < pc.VoiceType {
		pc.VoiceType = VoiceType_FM
		ok = false
	}
	switch pc.VoiceType {
	case VoiceType_FM:
		if pc.FmVoice == nil {
			pc.FmVoice = &VM35FMVoice{}
			ok = false
		}
		if !pc.FmVoice.Normalize() {
			ok = false
		}
	case VoiceType_PCM:
		if pc.PcmVoice == nil {
			pc.PcmVoice = &VM35PCMVoice{}
			ok = false
		}
		if !pc.PcmVoice.Normalize() {
			ok = false
		}
	case VoiceType_AL:
		pc.VoiceType = VoiceType_FM
		pc.FmVoice = &VM35FMVoice{}
		ok = false
	}
	return ok
}

// Normalize は、音色データから異常な値を排除し、正常化します。
// 元から正常な音色だったときは true を返します。
func (voice *VM35FMVoice) Normalize() bool {
	ok := true
	normalizeUint32(&ok, &voice.DrumKey, 0, 127)
	normalizeUint32(&ok, &voice.Panpot, 0, 31)
	normalizeUint32(&ok, &voice.Bo, 0, 3)
	normalizeUint32(&ok, &voice.Lfo, 0, 3)
	normalizeUint32(&ok, &voice.Alg, 0, 7)
	ops := 4
	if voice.Alg < 2 {
		ops = 2
	}
	for len(voice.Operators) < ops {
		voice.Operators = append(voice.Operators, &VM35FMOperator{})
		ok = false
	}
	for i, op := range voice.Operators {
		if op == nil {
			op = &VM35FMOperator{}
			voice.Operators[i] = op
			ok = false
		}
		if !op.Normalize() {
			ok = false
		}
	}
	return ok
}

// Normalize は、音色データから異常な値を排除し、正常化します。
// 元から正常な音色だったときは true を返します。
func (op *VM35FMOperator) Normalize() bool {
	ok := true
	normalizeUint32(&ok, &op.Multi, 0, 15)
	normalizeUint32(&ok, &op.Dt, 0, 7)
	normalizeUint32(&ok, &op.Ar, 0, 15)
	normalizeUint32(&ok, &op.Dr, 0, 15)
	normalizeUint32(&ok, &op.Sr, 0, 15)
	normalizeUint32(&ok, &op.Rr, 0, 15)
	normalizeUint32(&ok, &op.Sl, 0, 15)
	normalizeUint32(&ok, &op.Tl, 0, 31)
	normalizeUint32(&ok, &op.Ksl, 0, 3)
	normalizeUint32(&ok, &op.Dam, 0, 3)
	normalizeUint32(&ok, &op.Dvb, 0, 3)
	normalizeUint32(&ok, &op.Fb, 0, 7)
	normalizeUint32(&ok, &op.Ws, 0, 31)
	// TODO: ユーザ波形のwarning
	return ok
}

// Normalize は、音色データから異常な値を排除し、正常化します。
// 元から正常な音色だったときは true を返します。
func (voice *VM35PCMVoice) Normalize() bool {
	ok := true
	if voice.RawData == nil {
		voice.RawData = []byte{}
		ok = false
	}
	return ok
}
