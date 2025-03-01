// Code generated by protoc-gen-ts_proto. DO NOT EDIT.
// versions:
//   protoc-gen-ts_proto  v2.2.0
//   protoc               unknown
// source: store/export_archive.proto

/* eslint-disable */
import { BinaryReader, BinaryWriter } from "@bufbuild/protobuf/wire";
import Long from "long";
import { ExportFormat, exportFormatFromJSON, exportFormatToJSON, exportFormatToNumber } from "./common";

export const protobufPackage = "devsecdb.store";

export interface ExportArchivePayload {
  /** The exported file format. e.g. JSON, CSV, SQL */
  fileFormat: ExportFormat;
}

function createBaseExportArchivePayload(): ExportArchivePayload {
  return { fileFormat: ExportFormat.FORMAT_UNSPECIFIED };
}

export const ExportArchivePayload: MessageFns<ExportArchivePayload> = {
  encode(message: ExportArchivePayload, writer: BinaryWriter = new BinaryWriter()): BinaryWriter {
    if (message.fileFormat !== ExportFormat.FORMAT_UNSPECIFIED) {
      writer.uint32(8).int32(exportFormatToNumber(message.fileFormat));
    }
    return writer;
  },

  decode(input: BinaryReader | Uint8Array, length?: number): ExportArchivePayload {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseExportArchivePayload();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 8) {
            break;
          }

          message.fileFormat = exportFormatFromJSON(reader.int32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skip(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): ExportArchivePayload {
    return {
      fileFormat: isSet(object.fileFormat) ? exportFormatFromJSON(object.fileFormat) : ExportFormat.FORMAT_UNSPECIFIED,
    };
  },

  toJSON(message: ExportArchivePayload): unknown {
    const obj: any = {};
    if (message.fileFormat !== ExportFormat.FORMAT_UNSPECIFIED) {
      obj.fileFormat = exportFormatToJSON(message.fileFormat);
    }
    return obj;
  },

  create(base?: DeepPartial<ExportArchivePayload>): ExportArchivePayload {
    return ExportArchivePayload.fromPartial(base ?? {});
  },
  fromPartial(object: DeepPartial<ExportArchivePayload>): ExportArchivePayload {
    const message = createBaseExportArchivePayload();
    message.fileFormat = object.fileFormat ?? ExportFormat.FORMAT_UNSPECIFIED;
    return message;
  },
};

type Builtin = Date | Function | Uint8Array | string | number | boolean | undefined;

export type DeepPartial<T> = T extends Builtin ? T
  : T extends Long ? string | number | Long : T extends globalThis.Array<infer U> ? globalThis.Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>>
  : T extends {} ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;

function isSet(value: any): boolean {
  return value !== null && value !== undefined;
}

export interface MessageFns<T> {
  encode(message: T, writer?: BinaryWriter): BinaryWriter;
  decode(input: BinaryReader | Uint8Array, length?: number): T;
  fromJSON(object: any): T;
  toJSON(message: T): unknown;
  create(base?: DeepPartial<T>): T;
  fromPartial(object: DeepPartial<T>): T;
}
