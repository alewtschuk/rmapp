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
            const char *errorMsg = [[error localizedDescription] UTF8String];
            fprintf(stderr, "[rmapp] Failed to move %s to Trash\n", path);
        }

        return success;
    }
}