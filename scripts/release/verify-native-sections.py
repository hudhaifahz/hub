#!/usr/bin/env python3
"""Verify that removing embedded LLVM IR preserves native sections and symbols."""
import struct
import sys
from pathlib import Path


def members(path):
    data = Path(path).read_bytes()
    assert data[:8] == b"!<arch>\n", "Expected an ar archive"
    result = []
    cursor = 8
    while cursor < len(data):
        header = data[cursor:cursor + 60]
        assert header[58:60] == b"`\n", "Invalid archive header"
        size = int(header[48:58])
        body = data[cursor + 60:cursor + 60 + size]
        name = header[:16].decode().strip().rstrip("/")
        if name.startswith("#1/"):
            length = int(name[3:])
            name = body[:length].rstrip(b"\0").decode()
            body = body[length:]
        if not name.startswith("__.SYMDEF"):
            result.append((name, body))
        cursor += 60 + size + size % 2
    assert cursor == len(data), "Truncated archive"
    return result


def native_object(data):
    header = struct.unpack_from("<8I", data)
    assert header[0] == 0xFEEDFACF and header[3] == 1, "Expected a 64-bit Mach-O object"
    sections, names, commands, symbols = [], [None], [], None
    cursor = 32
    for _ in range(header[4]):
        command, size = struct.unpack_from("<II", data, cursor)
        assert size >= 8 and cursor + size <= len(data)
        if command == 0x19:  # LC_SEGMENT_64
            count = struct.unpack_from("<I", data, cursor + 64)[0]
            for index in range(count):
                section = struct.unpack_from("<16s16sQQIIIIIIII", data, cursor + 72 + index * 80)
                name, segment = (part.rstrip(b"\0") for part in section[:2])
                identity = (segment, name)
                names.append(identity)
                if segment == b"__LLVM" and name in (b"__bitcode", b"__cmdline"):
                    continue
                address, length, offset, align, reloc, count, flags, r1, r2, r3 = section[2:]
                contents = b"" if flags & 0xFF in (1, 0xC, 0x12) else data[offset:offset + length]
                sections.append((identity, address, length, align, flags, r1, r2, r3,
                                 contents, data[reloc:reloc + count * 8]))
        elif command == 2:  # LC_SYMTAB; offsets may move, symbol meaning must not.
            symbols = struct.unpack_from("<4I", data, cursor + 8)
        elif command in (0x29, 0x2E):  # DATA_IN_CODE / LINKER_OPTIMIZATION_HINT payload.
            offset, length = struct.unpack_from("<II", data, cursor + 8)
            commands.append((command, data[offset:offset + length]))
        elif command == 0xB:  # LC_DYSYMTAB indices change when LLVM-only symbols disappear.
            fields = struct.unpack_from("<18I", data, cursor + 8)
            assert not any(fields[6:]), "Unexpected indirect symbol/relocation table"
        else:
            commands.append(data[cursor:cursor + size])
        cursor += size
    normalized = []
    if symbols:
        offset, count, strings, _ = symbols
        for index in range(count):
            name, kind, section, desc, value = struct.unpack_from("<IBBHQ", data, offset + index * 16)
            end = data.index(b"\0", strings + name)
            normalized.append((data[strings + name:end], kind, names[section], desc, value))
    resolved_sections = []
    for section in sections:
        relocations = []
        for offset in range(0, len(section[-1]), 8):
            address, relocation = struct.unpack_from("<II", section[-1], offset)
            assert not address & 0x80000000, "Unexpected scattered relocation"
            index = relocation & 0xFFFFFF
            if relocation >> 28 == 10 and header[1] == 0x100000C:  # ARM64 ADDEND
                target = index
            elif relocation & (1 << 27):
                target = normalized[index]
            else:
                target = names[index]
            relocations.append((address, relocation >> 24, target))
        resolved_sections.append((*section[:-1], relocations))
    retained_symbols = [symbol for symbol in normalized if symbol[2] not in (
        (b"__LLVM", b"__bitcode"), (b"__LLVM", b"__cmdline"))]
    return (header[1:4], header[6:], resolved_sections, commands, retained_symbols)


def verify(original, transformed):
    before, after = members(original), members(transformed)
    assert len(before) == len(after), "Archive members changed"
    for (old_name, old), (new_name, new) in zip(before, after):
        assert old_name == new_name, "Archive member name/order changed"
        assert native_object(old) == native_object(new), f"Native contents changed: {old_name}"
        assert b"__bitcode" not in new[:struct.unpack_from("<I", new, 20)[0] + 32], "LLVM IR remains"
    print(f"Verified unchanged native sections, relocations and symbols in {len(before)} objects")


if __name__ == "__main__":
    verify(*sys.argv[1:])
