// Code generated by protoc-gen-ts_proto. DO NOT EDIT.
// versions:
//   protoc-gen-ts_proto  v2.2.0
//   protoc               unknown
// source: store/branch.proto

/* eslint-disable */
import { BinaryReader, BinaryWriter } from "@bufbuild/protobuf/wire";
import Long from "long";
import { DatabaseConfig, DatabaseSchemaMetadata } from "./database";

export const protobufPackage = "devsecdb.store";

export interface BranchSnapshot {
  metadata: DatabaseSchemaMetadata | undefined;
  databaseConfig: DatabaseConfig | undefined;
}

export interface BranchConfig {
  /**
   * The name of source database.
   * Optional.
   * Example: instances/instance-id/databases/database-name.
   */
  sourceDatabase: string;
  /**
   * The name of the source branch.
   * Optional.
   * Example: projects/project-id/branches/branch-id.
   */
  sourceBranch: string;
}

function createBaseBranchSnapshot(): BranchSnapshot {
  return { metadata: undefined, databaseConfig: undefined };
}

export const BranchSnapshot: MessageFns<BranchSnapshot> = {
  encode(message: BranchSnapshot, writer: BinaryWriter = new BinaryWriter()): BinaryWriter {
    if (message.metadata !== undefined) {
      DatabaseSchemaMetadata.encode(message.metadata, writer.uint32(10).fork()).join();
    }
    if (message.databaseConfig !== undefined) {
      DatabaseConfig.encode(message.databaseConfig, writer.uint32(18).fork()).join();
    }
    return writer;
  },

  decode(input: BinaryReader | Uint8Array, length?: number): BranchSnapshot {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseBranchSnapshot();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.metadata = DatabaseSchemaMetadata.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.databaseConfig = DatabaseConfig.decode(reader, reader.uint32());
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skip(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): BranchSnapshot {
    return {
      metadata: isSet(object.metadata) ? DatabaseSchemaMetadata.fromJSON(object.metadata) : undefined,
      databaseConfig: isSet(object.databaseConfig) ? DatabaseConfig.fromJSON(object.databaseConfig) : undefined,
    };
  },

  toJSON(message: BranchSnapshot): unknown {
    const obj: any = {};
    if (message.metadata !== undefined) {
      obj.metadata = DatabaseSchemaMetadata.toJSON(message.metadata);
    }
    if (message.databaseConfig !== undefined) {
      obj.databaseConfig = DatabaseConfig.toJSON(message.databaseConfig);
    }
    return obj;
  },

  create(base?: DeepPartial<BranchSnapshot>): BranchSnapshot {
    return BranchSnapshot.fromPartial(base ?? {});
  },
  fromPartial(object: DeepPartial<BranchSnapshot>): BranchSnapshot {
    const message = createBaseBranchSnapshot();
    message.metadata = (object.metadata !== undefined && object.metadata !== null)
      ? DatabaseSchemaMetadata.fromPartial(object.metadata)
      : undefined;
    message.databaseConfig = (object.databaseConfig !== undefined && object.databaseConfig !== null)
      ? DatabaseConfig.fromPartial(object.databaseConfig)
      : undefined;
    return message;
  },
};

function createBaseBranchConfig(): BranchConfig {
  return { sourceDatabase: "", sourceBranch: "" };
}

export const BranchConfig: MessageFns<BranchConfig> = {
  encode(message: BranchConfig, writer: BinaryWriter = new BinaryWriter()): BinaryWriter {
    if (message.sourceDatabase !== "") {
      writer.uint32(10).string(message.sourceDatabase);
    }
    if (message.sourceBranch !== "") {
      writer.uint32(18).string(message.sourceBranch);
    }
    return writer;
  },

  decode(input: BinaryReader | Uint8Array, length?: number): BranchConfig {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseBranchConfig();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.sourceDatabase = reader.string();
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.sourceBranch = reader.string();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skip(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): BranchConfig {
    return {
      sourceDatabase: isSet(object.sourceDatabase) ? globalThis.String(object.sourceDatabase) : "",
      sourceBranch: isSet(object.sourceBranch) ? globalThis.String(object.sourceBranch) : "",
    };
  },

  toJSON(message: BranchConfig): unknown {
    const obj: any = {};
    if (message.sourceDatabase !== "") {
      obj.sourceDatabase = message.sourceDatabase;
    }
    if (message.sourceBranch !== "") {
      obj.sourceBranch = message.sourceBranch;
    }
    return obj;
  },

  create(base?: DeepPartial<BranchConfig>): BranchConfig {
    return BranchConfig.fromPartial(base ?? {});
  },
  fromPartial(object: DeepPartial<BranchConfig>): BranchConfig {
    const message = createBaseBranchConfig();
    message.sourceDatabase = object.sourceDatabase ?? "";
    message.sourceBranch = object.sourceBranch ?? "";
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
