export type TryCatchResult<T> =
  | { value: T; error?: never }
  | { value?: never; error: Error };

export async function tryCatch<T>(cb: Promise<T>): Promise<TryCatchResult<T>> {
  try {
    return { value: await cb };
  } catch (error) {
    return {
      error:
        error instanceof Error
          ? error
          : new Error("unknown error", { cause: error }),
    };
  }
}
