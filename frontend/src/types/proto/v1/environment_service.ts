// Code generated by protoc-gen-ts_proto. DO NOT EDIT.
// versions:
//   protoc-gen-ts_proto  v2.2.0
//   protoc               unknown
// source: v1/environment_service.proto

/* eslint-disable */
import { BinaryReader, BinaryWriter } from "@bufbuild/protobuf/wire";
import Long from "long";
import { Empty } from "../google/protobuf/empty";
import { FieldMask } from "../google/protobuf/field_mask";
import { State, stateFromJSON, stateToJSON, stateToNumber } from "./common";

export const protobufPackage = "devsecdb.v1";

export enum EnvironmentTier {
  ENVIRONMENT_TIER_UNSPECIFIED = "ENVIRONMENT_TIER_UNSPECIFIED",
  PROTECTED = "PROTECTED",
  UNPROTECTED = "UNPROTECTED",
  UNRECOGNIZED = "UNRECOGNIZED",
}

export function environmentTierFromJSON(object: any): EnvironmentTier {
  switch (object) {
    case 0:
    case "ENVIRONMENT_TIER_UNSPECIFIED":
      return EnvironmentTier.ENVIRONMENT_TIER_UNSPECIFIED;
    case 1:
    case "PROTECTED":
      return EnvironmentTier.PROTECTED;
    case 2:
    case "UNPROTECTED":
      return EnvironmentTier.UNPROTECTED;
    case -1:
    case "UNRECOGNIZED":
    default:
      return EnvironmentTier.UNRECOGNIZED;
  }
}

export function environmentTierToJSON(object: EnvironmentTier): string {
  switch (object) {
    case EnvironmentTier.ENVIRONMENT_TIER_UNSPECIFIED:
      return "ENVIRONMENT_TIER_UNSPECIFIED";
    case EnvironmentTier.PROTECTED:
      return "PROTECTED";
    case EnvironmentTier.UNPROTECTED:
      return "UNPROTECTED";
    case EnvironmentTier.UNRECOGNIZED:
    default:
      return "UNRECOGNIZED";
  }
}

export function environmentTierToNumber(object: EnvironmentTier): number {
  switch (object) {
    case EnvironmentTier.ENVIRONMENT_TIER_UNSPECIFIED:
      return 0;
    case EnvironmentTier.PROTECTED:
      return 1;
    case EnvironmentTier.UNPROTECTED:
      return 2;
    case EnvironmentTier.UNRECOGNIZED:
    default:
      return -1;
  }
}

export interface GetEnvironmentRequest {
  /**
   * The name of the environment to retrieve.
   * Format: environments/{environment}
   */
  name: string;
}

export interface ListEnvironmentsRequest {
  /**
   * Not used.
   * The maximum number of environments to return. The service may return fewer than
   * this value.
   * If unspecified, at most 10 environments will be returned.
   * The maximum value is 1000; values above 1000 will be coerced to 1000.
   */
  pageSize: number;
  /**
   * Not used.
   * A page token, received from a previous `ListEnvironments` call.
   * Provide this to retrieve the subsequent page.
   *
   * When paginating, all other parameters provided to `ListEnvironments` must match
   * the call that provided the page token.
   */
  pageToken: string;
  /** Show deleted environments if specified. */
  showDeleted: boolean;
}

export interface ListEnvironmentsResponse {
  /** The environments from the specified request. */
  environments: Environment[];
  /**
   * A token, which can be sent as `page_token` to retrieve the next page.
   * If this field is omitted, there are no subsequent pages.
   */
  nextPageToken: string;
}

export interface CreateEnvironmentRequest {
  /** The environment to create. */
  environment:
    | Environment
    | undefined;
  /**
   * The ID to use for the environment, which will become the final component of
   * the environment's resource name.
   *
   * This value should be 4-63 characters, and valid characters
   * are /[a-z][0-9]-/.
   */
  environmentId: string;
}

export interface UpdateEnvironmentRequest {
  /**
   * The environment to update.
   *
   * The environment's `name` field is used to identify the environment to update.
   * Format: environments/{environment}
   */
  environment:
    | Environment
    | undefined;
  /** The list of fields to update. */
  updateMask: string[] | undefined;
}

export interface DeleteEnvironmentRequest {
  /**
   * The name of the environment to delete.
   * Format: environments/{environment}
   */
  name: string;
}

export interface UndeleteEnvironmentRequest {
  /**
   * The name of the deleted environment.
   * Format: environments/{environment}
   */
  name: string;
}

export interface Environment {
  /**
   * The name of the environment.
   * Format: environments/{environment}
   */
  name: string;
  state: State;
  title: string;
  order: number;
  tier: EnvironmentTier;
  color: string;
}

function createBaseGetEnvironmentRequest(): GetEnvironmentRequest {
  return { name: "" };
}

export const GetEnvironmentRequest: MessageFns<GetEnvironmentRequest> = {
  encode(message: GetEnvironmentRequest, writer: BinaryWriter = new BinaryWriter()): BinaryWriter {
    if (message.name !== "") {
      writer.uint32(10).string(message.name);
    }
    return writer;
  },

  decode(input: BinaryReader | Uint8Array, length?: number): GetEnvironmentRequest {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseGetEnvironmentRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.name = reader.string();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skip(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): GetEnvironmentRequest {
    return { name: isSet(object.name) ? globalThis.String(object.name) : "" };
  },

  toJSON(message: GetEnvironmentRequest): unknown {
    const obj: any = {};
    if (message.name !== "") {
      obj.name = message.name;
    }
    return obj;
  },

  create(base?: DeepPartial<GetEnvironmentRequest>): GetEnvironmentRequest {
    return GetEnvironmentRequest.fromPartial(base ?? {});
  },
  fromPartial(object: DeepPartial<GetEnvironmentRequest>): GetEnvironmentRequest {
    const message = createBaseGetEnvironmentRequest();
    message.name = object.name ?? "";
    return message;
  },
};

function createBaseListEnvironmentsRequest(): ListEnvironmentsRequest {
  return { pageSize: 0, pageToken: "", showDeleted: false };
}

export const ListEnvironmentsRequest: MessageFns<ListEnvironmentsRequest> = {
  encode(message: ListEnvironmentsRequest, writer: BinaryWriter = new BinaryWriter()): BinaryWriter {
    if (message.pageSize !== 0) {
      writer.uint32(8).int32(message.pageSize);
    }
    if (message.pageToken !== "") {
      writer.uint32(18).string(message.pageToken);
    }
    if (message.showDeleted !== false) {
      writer.uint32(24).bool(message.showDeleted);
    }
    return writer;
  },

  decode(input: BinaryReader | Uint8Array, length?: number): ListEnvironmentsRequest {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseListEnvironmentsRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 8) {
            break;
          }

          message.pageSize = reader.int32();
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.pageToken = reader.string();
          continue;
        case 3:
          if (tag !== 24) {
            break;
          }

          message.showDeleted = reader.bool();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skip(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): ListEnvironmentsRequest {
    return {
      pageSize: isSet(object.pageSize) ? globalThis.Number(object.pageSize) : 0,
      pageToken: isSet(object.pageToken) ? globalThis.String(object.pageToken) : "",
      showDeleted: isSet(object.showDeleted) ? globalThis.Boolean(object.showDeleted) : false,
    };
  },

  toJSON(message: ListEnvironmentsRequest): unknown {
    const obj: any = {};
    if (message.pageSize !== 0) {
      obj.pageSize = Math.round(message.pageSize);
    }
    if (message.pageToken !== "") {
      obj.pageToken = message.pageToken;
    }
    if (message.showDeleted !== false) {
      obj.showDeleted = message.showDeleted;
    }
    return obj;
  },

  create(base?: DeepPartial<ListEnvironmentsRequest>): ListEnvironmentsRequest {
    return ListEnvironmentsRequest.fromPartial(base ?? {});
  },
  fromPartial(object: DeepPartial<ListEnvironmentsRequest>): ListEnvironmentsRequest {
    const message = createBaseListEnvironmentsRequest();
    message.pageSize = object.pageSize ?? 0;
    message.pageToken = object.pageToken ?? "";
    message.showDeleted = object.showDeleted ?? false;
    return message;
  },
};

function createBaseListEnvironmentsResponse(): ListEnvironmentsResponse {
  return { environments: [], nextPageToken: "" };
}

export const ListEnvironmentsResponse: MessageFns<ListEnvironmentsResponse> = {
  encode(message: ListEnvironmentsResponse, writer: BinaryWriter = new BinaryWriter()): BinaryWriter {
    for (const v of message.environments) {
      Environment.encode(v!, writer.uint32(10).fork()).join();
    }
    if (message.nextPageToken !== "") {
      writer.uint32(18).string(message.nextPageToken);
    }
    return writer;
  },

  decode(input: BinaryReader | Uint8Array, length?: number): ListEnvironmentsResponse {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseListEnvironmentsResponse();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.environments.push(Environment.decode(reader, reader.uint32()));
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.nextPageToken = reader.string();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skip(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): ListEnvironmentsResponse {
    return {
      environments: globalThis.Array.isArray(object?.environments)
        ? object.environments.map((e: any) => Environment.fromJSON(e))
        : [],
      nextPageToken: isSet(object.nextPageToken) ? globalThis.String(object.nextPageToken) : "",
    };
  },

  toJSON(message: ListEnvironmentsResponse): unknown {
    const obj: any = {};
    if (message.environments?.length) {
      obj.environments = message.environments.map((e) => Environment.toJSON(e));
    }
    if (message.nextPageToken !== "") {
      obj.nextPageToken = message.nextPageToken;
    }
    return obj;
  },

  create(base?: DeepPartial<ListEnvironmentsResponse>): ListEnvironmentsResponse {
    return ListEnvironmentsResponse.fromPartial(base ?? {});
  },
  fromPartial(object: DeepPartial<ListEnvironmentsResponse>): ListEnvironmentsResponse {
    const message = createBaseListEnvironmentsResponse();
    message.environments = object.environments?.map((e) => Environment.fromPartial(e)) || [];
    message.nextPageToken = object.nextPageToken ?? "";
    return message;
  },
};

function createBaseCreateEnvironmentRequest(): CreateEnvironmentRequest {
  return { environment: undefined, environmentId: "" };
}

export const CreateEnvironmentRequest: MessageFns<CreateEnvironmentRequest> = {
  encode(message: CreateEnvironmentRequest, writer: BinaryWriter = new BinaryWriter()): BinaryWriter {
    if (message.environment !== undefined) {
      Environment.encode(message.environment, writer.uint32(10).fork()).join();
    }
    if (message.environmentId !== "") {
      writer.uint32(18).string(message.environmentId);
    }
    return writer;
  },

  decode(input: BinaryReader | Uint8Array, length?: number): CreateEnvironmentRequest {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseCreateEnvironmentRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.environment = Environment.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.environmentId = reader.string();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skip(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): CreateEnvironmentRequest {
    return {
      environment: isSet(object.environment) ? Environment.fromJSON(object.environment) : undefined,
      environmentId: isSet(object.environmentId) ? globalThis.String(object.environmentId) : "",
    };
  },

  toJSON(message: CreateEnvironmentRequest): unknown {
    const obj: any = {};
    if (message.environment !== undefined) {
      obj.environment = Environment.toJSON(message.environment);
    }
    if (message.environmentId !== "") {
      obj.environmentId = message.environmentId;
    }
    return obj;
  },

  create(base?: DeepPartial<CreateEnvironmentRequest>): CreateEnvironmentRequest {
    return CreateEnvironmentRequest.fromPartial(base ?? {});
  },
  fromPartial(object: DeepPartial<CreateEnvironmentRequest>): CreateEnvironmentRequest {
    const message = createBaseCreateEnvironmentRequest();
    message.environment = (object.environment !== undefined && object.environment !== null)
      ? Environment.fromPartial(object.environment)
      : undefined;
    message.environmentId = object.environmentId ?? "";
    return message;
  },
};

function createBaseUpdateEnvironmentRequest(): UpdateEnvironmentRequest {
  return { environment: undefined, updateMask: undefined };
}

export const UpdateEnvironmentRequest: MessageFns<UpdateEnvironmentRequest> = {
  encode(message: UpdateEnvironmentRequest, writer: BinaryWriter = new BinaryWriter()): BinaryWriter {
    if (message.environment !== undefined) {
      Environment.encode(message.environment, writer.uint32(10).fork()).join();
    }
    if (message.updateMask !== undefined) {
      FieldMask.encode(FieldMask.wrap(message.updateMask), writer.uint32(18).fork()).join();
    }
    return writer;
  },

  decode(input: BinaryReader | Uint8Array, length?: number): UpdateEnvironmentRequest {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseUpdateEnvironmentRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.environment = Environment.decode(reader, reader.uint32());
          continue;
        case 2:
          if (tag !== 18) {
            break;
          }

          message.updateMask = FieldMask.unwrap(FieldMask.decode(reader, reader.uint32()));
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skip(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): UpdateEnvironmentRequest {
    return {
      environment: isSet(object.environment) ? Environment.fromJSON(object.environment) : undefined,
      updateMask: isSet(object.updateMask) ? FieldMask.unwrap(FieldMask.fromJSON(object.updateMask)) : undefined,
    };
  },

  toJSON(message: UpdateEnvironmentRequest): unknown {
    const obj: any = {};
    if (message.environment !== undefined) {
      obj.environment = Environment.toJSON(message.environment);
    }
    if (message.updateMask !== undefined) {
      obj.updateMask = FieldMask.toJSON(FieldMask.wrap(message.updateMask));
    }
    return obj;
  },

  create(base?: DeepPartial<UpdateEnvironmentRequest>): UpdateEnvironmentRequest {
    return UpdateEnvironmentRequest.fromPartial(base ?? {});
  },
  fromPartial(object: DeepPartial<UpdateEnvironmentRequest>): UpdateEnvironmentRequest {
    const message = createBaseUpdateEnvironmentRequest();
    message.environment = (object.environment !== undefined && object.environment !== null)
      ? Environment.fromPartial(object.environment)
      : undefined;
    message.updateMask = object.updateMask ?? undefined;
    return message;
  },
};

function createBaseDeleteEnvironmentRequest(): DeleteEnvironmentRequest {
  return { name: "" };
}

export const DeleteEnvironmentRequest: MessageFns<DeleteEnvironmentRequest> = {
  encode(message: DeleteEnvironmentRequest, writer: BinaryWriter = new BinaryWriter()): BinaryWriter {
    if (message.name !== "") {
      writer.uint32(10).string(message.name);
    }
    return writer;
  },

  decode(input: BinaryReader | Uint8Array, length?: number): DeleteEnvironmentRequest {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseDeleteEnvironmentRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.name = reader.string();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skip(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): DeleteEnvironmentRequest {
    return { name: isSet(object.name) ? globalThis.String(object.name) : "" };
  },

  toJSON(message: DeleteEnvironmentRequest): unknown {
    const obj: any = {};
    if (message.name !== "") {
      obj.name = message.name;
    }
    return obj;
  },

  create(base?: DeepPartial<DeleteEnvironmentRequest>): DeleteEnvironmentRequest {
    return DeleteEnvironmentRequest.fromPartial(base ?? {});
  },
  fromPartial(object: DeepPartial<DeleteEnvironmentRequest>): DeleteEnvironmentRequest {
    const message = createBaseDeleteEnvironmentRequest();
    message.name = object.name ?? "";
    return message;
  },
};

function createBaseUndeleteEnvironmentRequest(): UndeleteEnvironmentRequest {
  return { name: "" };
}

export const UndeleteEnvironmentRequest: MessageFns<UndeleteEnvironmentRequest> = {
  encode(message: UndeleteEnvironmentRequest, writer: BinaryWriter = new BinaryWriter()): BinaryWriter {
    if (message.name !== "") {
      writer.uint32(10).string(message.name);
    }
    return writer;
  },

  decode(input: BinaryReader | Uint8Array, length?: number): UndeleteEnvironmentRequest {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseUndeleteEnvironmentRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.name = reader.string();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skip(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): UndeleteEnvironmentRequest {
    return { name: isSet(object.name) ? globalThis.String(object.name) : "" };
  },

  toJSON(message: UndeleteEnvironmentRequest): unknown {
    const obj: any = {};
    if (message.name !== "") {
      obj.name = message.name;
    }
    return obj;
  },

  create(base?: DeepPartial<UndeleteEnvironmentRequest>): UndeleteEnvironmentRequest {
    return UndeleteEnvironmentRequest.fromPartial(base ?? {});
  },
  fromPartial(object: DeepPartial<UndeleteEnvironmentRequest>): UndeleteEnvironmentRequest {
    const message = createBaseUndeleteEnvironmentRequest();
    message.name = object.name ?? "";
    return message;
  },
};

function createBaseEnvironment(): Environment {
  return {
    name: "",
    state: State.STATE_UNSPECIFIED,
    title: "",
    order: 0,
    tier: EnvironmentTier.ENVIRONMENT_TIER_UNSPECIFIED,
    color: "",
  };
}

export const Environment: MessageFns<Environment> = {
  encode(message: Environment, writer: BinaryWriter = new BinaryWriter()): BinaryWriter {
    if (message.name !== "") {
      writer.uint32(10).string(message.name);
    }
    if (message.state !== State.STATE_UNSPECIFIED) {
      writer.uint32(24).int32(stateToNumber(message.state));
    }
    if (message.title !== "") {
      writer.uint32(34).string(message.title);
    }
    if (message.order !== 0) {
      writer.uint32(40).int32(message.order);
    }
    if (message.tier !== EnvironmentTier.ENVIRONMENT_TIER_UNSPECIFIED) {
      writer.uint32(48).int32(environmentTierToNumber(message.tier));
    }
    if (message.color !== "") {
      writer.uint32(58).string(message.color);
    }
    return writer;
  },

  decode(input: BinaryReader | Uint8Array, length?: number): Environment {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseEnvironment();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break;
          }

          message.name = reader.string();
          continue;
        case 3:
          if (tag !== 24) {
            break;
          }

          message.state = stateFromJSON(reader.int32());
          continue;
        case 4:
          if (tag !== 34) {
            break;
          }

          message.title = reader.string();
          continue;
        case 5:
          if (tag !== 40) {
            break;
          }

          message.order = reader.int32();
          continue;
        case 6:
          if (tag !== 48) {
            break;
          }

          message.tier = environmentTierFromJSON(reader.int32());
          continue;
        case 7:
          if (tag !== 58) {
            break;
          }

          message.color = reader.string();
          continue;
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skip(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): Environment {
    return {
      name: isSet(object.name) ? globalThis.String(object.name) : "",
      state: isSet(object.state) ? stateFromJSON(object.state) : State.STATE_UNSPECIFIED,
      title: isSet(object.title) ? globalThis.String(object.title) : "",
      order: isSet(object.order) ? globalThis.Number(object.order) : 0,
      tier: isSet(object.tier) ? environmentTierFromJSON(object.tier) : EnvironmentTier.ENVIRONMENT_TIER_UNSPECIFIED,
      color: isSet(object.color) ? globalThis.String(object.color) : "",
    };
  },

  toJSON(message: Environment): unknown {
    const obj: any = {};
    if (message.name !== "") {
      obj.name = message.name;
    }
    if (message.state !== State.STATE_UNSPECIFIED) {
      obj.state = stateToJSON(message.state);
    }
    if (message.title !== "") {
      obj.title = message.title;
    }
    if (message.order !== 0) {
      obj.order = Math.round(message.order);
    }
    if (message.tier !== EnvironmentTier.ENVIRONMENT_TIER_UNSPECIFIED) {
      obj.tier = environmentTierToJSON(message.tier);
    }
    if (message.color !== "") {
      obj.color = message.color;
    }
    return obj;
  },

  create(base?: DeepPartial<Environment>): Environment {
    return Environment.fromPartial(base ?? {});
  },
  fromPartial(object: DeepPartial<Environment>): Environment {
    const message = createBaseEnvironment();
    message.name = object.name ?? "";
    message.state = object.state ?? State.STATE_UNSPECIFIED;
    message.title = object.title ?? "";
    message.order = object.order ?? 0;
    message.tier = object.tier ?? EnvironmentTier.ENVIRONMENT_TIER_UNSPECIFIED;
    message.color = object.color ?? "";
    return message;
  },
};

export type EnvironmentServiceDefinition = typeof EnvironmentServiceDefinition;
export const EnvironmentServiceDefinition = {
  name: "EnvironmentService",
  fullName: "devsecdb.v1.EnvironmentService",
  methods: {
    getEnvironment: {
      name: "GetEnvironment",
      requestType: GetEnvironmentRequest,
      requestStream: false,
      responseType: Environment,
      responseStream: false,
      options: {
        _unknownFields: {
          8410: [new Uint8Array([4, 110, 97, 109, 101])],
          800010: [
            new Uint8Array([
              19,
              98,
              98,
              46,
              101,
              110,
              118,
              105,
              114,
              111,
              110,
              109,
              101,
              110,
              116,
              115,
              46,
              103,
              101,
              116,
            ]),
          ],
          800016: [new Uint8Array([1])],
          578365826: [
            new Uint8Array([
              27,
              18,
              25,
              47,
              118,
              49,
              47,
              123,
              110,
              97,
              109,
              101,
              61,
              101,
              110,
              118,
              105,
              114,
              111,
              110,
              109,
              101,
              110,
              116,
              115,
              47,
              42,
              125,
            ]),
          ],
        },
      },
    },
    listEnvironments: {
      name: "ListEnvironments",
      requestType: ListEnvironmentsRequest,
      requestStream: false,
      responseType: ListEnvironmentsResponse,
      responseStream: false,
      options: {
        _unknownFields: {
          8410: [new Uint8Array([0])],
          800010: [
            new Uint8Array([
              20,
              98,
              98,
              46,
              101,
              110,
              118,
              105,
              114,
              111,
              110,
              109,
              101,
              110,
              116,
              115,
              46,
              108,
              105,
              115,
              116,
            ]),
          ],
          800016: [new Uint8Array([1])],
          578365826: [
            new Uint8Array([18, 18, 16, 47, 118, 49, 47, 101, 110, 118, 105, 114, 111, 110, 109, 101, 110, 116, 115]),
          ],
        },
      },
    },
    createEnvironment: {
      name: "CreateEnvironment",
      requestType: CreateEnvironmentRequest,
      requestStream: false,
      responseType: Environment,
      responseStream: false,
      options: {
        _unknownFields: {
          8410: [new Uint8Array([0])],
          800010: [
            new Uint8Array([
              22,
              98,
              98,
              46,
              101,
              110,
              118,
              105,
              114,
              111,
              110,
              109,
              101,
              110,
              116,
              115,
              46,
              99,
              114,
              101,
              97,
              116,
              101,
            ]),
          ],
          800016: [new Uint8Array([1])],
          800024: [new Uint8Array([1])],
          578365826: [
            new Uint8Array([
              31,
              58,
              11,
              101,
              110,
              118,
              105,
              114,
              111,
              110,
              109,
              101,
              110,
              116,
              34,
              16,
              47,
              118,
              49,
              47,
              101,
              110,
              118,
              105,
              114,
              111,
              110,
              109,
              101,
              110,
              116,
              115,
            ]),
          ],
        },
      },
    },
    updateEnvironment: {
      name: "UpdateEnvironment",
      requestType: UpdateEnvironmentRequest,
      requestStream: false,
      responseType: Environment,
      responseStream: false,
      options: {
        _unknownFields: {
          8410: [
            new Uint8Array([
              23,
              101,
              110,
              118,
              105,
              114,
              111,
              110,
              109,
              101,
              110,
              116,
              44,
              117,
              112,
              100,
              97,
              116,
              101,
              95,
              109,
              97,
              115,
              107,
            ]),
          ],
          800010: [
            new Uint8Array([
              22,
              98,
              98,
              46,
              101,
              110,
              118,
              105,
              114,
              111,
              110,
              109,
              101,
              110,
              116,
              115,
              46,
              117,
              112,
              100,
              97,
              116,
              101,
            ]),
          ],
          800016: [new Uint8Array([1])],
          800024: [new Uint8Array([1])],
          578365826: [
            new Uint8Array([
              52,
              58,
              11,
              101,
              110,
              118,
              105,
              114,
              111,
              110,
              109,
              101,
              110,
              116,
              50,
              37,
              47,
              118,
              49,
              47,
              123,
              101,
              110,
              118,
              105,
              114,
              111,
              110,
              109,
              101,
              110,
              116,
              46,
              110,
              97,
              109,
              101,
              61,
              101,
              110,
              118,
              105,
              114,
              111,
              110,
              109,
              101,
              110,
              116,
              115,
              47,
              42,
              125,
            ]),
          ],
        },
      },
    },
    deleteEnvironment: {
      name: "DeleteEnvironment",
      requestType: DeleteEnvironmentRequest,
      requestStream: false,
      responseType: Empty,
      responseStream: false,
      options: {
        _unknownFields: {
          8410: [new Uint8Array([4, 110, 97, 109, 101])],
          800010: [
            new Uint8Array([
              22,
              98,
              98,
              46,
              101,
              110,
              118,
              105,
              114,
              111,
              110,
              109,
              101,
              110,
              116,
              115,
              46,
              100,
              101,
              108,
              101,
              116,
              101,
            ]),
          ],
          800016: [new Uint8Array([1])],
          800024: [new Uint8Array([1])],
          578365826: [
            new Uint8Array([
              27,
              42,
              25,
              47,
              118,
              49,
              47,
              123,
              110,
              97,
              109,
              101,
              61,
              101,
              110,
              118,
              105,
              114,
              111,
              110,
              109,
              101,
              110,
              116,
              115,
              47,
              42,
              125,
            ]),
          ],
        },
      },
    },
    undeleteEnvironment: {
      name: "UndeleteEnvironment",
      requestType: UndeleteEnvironmentRequest,
      requestStream: false,
      responseType: Environment,
      responseStream: false,
      options: {
        _unknownFields: {
          800010: [
            new Uint8Array([
              24,
              98,
              98,
              46,
              101,
              110,
              118,
              105,
              114,
              111,
              110,
              109,
              101,
              110,
              116,
              115,
              46,
              117,
              110,
              100,
              101,
              108,
              101,
              116,
              101,
            ]),
          ],
          800016: [new Uint8Array([1])],
          800024: [new Uint8Array([1])],
          578365826: [
            new Uint8Array([
              39,
              58,
              1,
              42,
              34,
              34,
              47,
              118,
              49,
              47,
              123,
              110,
              97,
              109,
              101,
              61,
              101,
              110,
              118,
              105,
              114,
              111,
              110,
              109,
              101,
              110,
              116,
              115,
              47,
              42,
              125,
              58,
              117,
              110,
              100,
              101,
              108,
              101,
              116,
              101,
            ]),
          ],
        },
      },
    },
  },
} as const;

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
