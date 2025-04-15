// Trash.mm - Serves as a native objective C wrapper to access MacOS NSFileManager

#import <Foundation/Foundation.h>

bool MoveToTrash(const char *path) {
    @autoreleasepool {
        NSString *filePath = [NSString stringWithUTF8String:path];
        NSURL *fileURL = [NSURL fileURLWithPath:filePath];
        NSError *error = nil;

        BOOL success = [[NSFileManager defaultManager]
                        trashItemAtURL:fileURL
                        resultingItemURL:nil
                        error:&error];

        if (!success) {
            NSLog(@"[rmapp] Failed to move %@ to Trash: %@", filePath, error);
        }

        return success;
    }
}