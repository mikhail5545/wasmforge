/*
 * Copyright (c) 2026. Mikhail Kulik.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React from "react"
import { AnimatePresence, motion } from "motion/react"
import { cva } from "class-variance-authority"
import { cn } from "@workspace/ui/lib/utils"

interface AlertModalProps {
  title: string
  description: string
  visible: boolean
  variant: "default" | "alert" | null | undefined
  size: "default" | "sm" | "md" | "lg" | null | undefined
  onClose: () => void
  icon?: React.JSX.Element
  timeout?: number
  className?: string
}

const alertVariants = cva(
  "fixed top-2 left-1/3 flex w-full flex-row gap-1 rounded-xl border border-border bg-background p-3 text-foreground md:w-1/2 lg:w-1/3",
  {
    variants: {
      variant: {
        default: "border-border bg-background text-foreground",
        alert: "border-border bg-destructive text-destructive-foreground",
      },
      size: {
        default: "left-0 left-1/3 w-full md:left-1/2 md:w-1/2 lg:w-1/3",
        sm: "left-1/3 w-1/3",
        md: "left-1/2 w-1/2",
        lg: "left-0 w-full",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

const AlertModal = ({
  title,
  description,
  visible,
  icon,
  variant = "default",
  size = "default",
  onClose,
  timeout = 5,
  className
}: AlertModalProps) => {
  const timeoutRef = React.useRef<NodeJS.Timeout | null>(null)

  React.useEffect(() => {
    if (visible) {
      timeoutRef.current = setTimeout(() => {
        timeoutRef.current = null
        onClose()
      }, timeout * 1000)
    } else if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
      timeoutRef.current = null
    }

    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
        timeoutRef.current = null
      }
    }
  }, [visible, timeout, onClose])

  return (
    <AnimatePresence mode={"popLayout"}>
      {visible && (
        <motion.div
          key={"alert"}
          initial={{ opacity: 0, scale: 0.5, y: -200 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          exit={{ opacity: 0, scale: 0.5, y: -200 }}
          transition={{ duration: 0.5, ease: "easeIn", type: "spring" }}
          className={cn(alertVariants({ size, variant, className }))}
        >
          <div className={"flex items-start justify-start p-2"}>{icon}</div>
          <div className={"flex flex-col gap-2"}>
            <div className={"flex items-center justify-start text-start"}>
              {title}
            </div>
            <div className={"text-left text-sm"}>
              {description}
            </div>
          </div>
        </motion.div>
      )}
    </AnimatePresence>
  )
}
AlertModal.displayName = "AlertModal"

export { AlertModal }
