import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

/** shadcn's class joiner: clsx for conditions, tailwind-merge so a later utility wins over an earlier one. */
export const cn = (...inputs: ClassValue[]): string => twMerge(clsx(inputs));
