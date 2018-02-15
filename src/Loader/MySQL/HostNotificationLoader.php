<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2018  Daniel Ziegler
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

namespace Statusengine\Loader\Mysql;

use Statusengine\Backend\StorageBackend;
use Statusengine\Loader\HostNotificationLoaderInterface;
use Statusengine\ValueObjects\HostNotificationQueryOptions;

class HostNotificationLoader implements HostNotificationLoaderInterface {


    /**
     * @var \Statusengine\Backend\Mysql\MySQL
     */
    private $Backend;

    /**
     * HostNotificationLoader constructor.
     * @param StorageBackend $StorageBackend
     */
    public function __construct(StorageBackend $StorageBackend) {
        $this->Backend = $StorageBackend->getBackend();
    }

    /**
     * @param HostNotificationQueryOptions $HostNotificationQueryOptions
     * @return array
     */
    public function getHostNotifications(HostNotificationQueryOptions $HostNotificationQueryOptions) {
        $fields = [
            'strings' => [
                'hostname',
                'output',
                'state',
                'command_name',
                'contact_name',
                'reason_type',
                'start_time'
            ]
        ];

        $sql = [];
        foreach ($fields['strings'] as $field) {
            $sql[] = $field;
        }


        $baseQuery = sprintf('SELECT %s FROM statusengine_host_notifications WHERE hostname=?', implode(',', $sql));

        if ($HostNotificationQueryOptions->sizeOfStateFilter() > 0 && $HostNotificationQueryOptions->sizeOfStateFilter() < 3) {
            $baseQuery = sprintf('%s AND state IN(%s)', $baseQuery, implode(',', $HostNotificationQueryOptions->getStateFilter()));
        }

        if ($HostNotificationQueryOptions->getOutputLike() != '') {
            $baseQuery = sprintf(' %s AND output LIKE ? ', $baseQuery);
        }

        if ($HostNotificationQueryOptions->hasReasonTypeFilter()) {
            $baseQuery = sprintf(' %s reason_type=?', $baseQuery);
        }

        $baseQuery = sprintf(
            '%s ORDER BY %s %s LIMIT ? OFFSET ?',
            $baseQuery,
            $HostNotificationQueryOptions->getOrder(),
            $HostNotificationQueryOptions->getDirection()
        );

        $query = $this->Backend->prepare($baseQuery);

        $i = 1;
        $query->bindValue($i++, $HostNotificationQueryOptions->getHostname());

        if ($HostNotificationQueryOptions->hasReasonTypeFilter()) {
            $query->bindValue($i++, $HostNotificationQueryOptions->getReasonType());
        }

        if ($HostNotificationQueryOptions->getOutputLike() != '') {
            $like = sprintf('%%%s%%', $HostNotificationQueryOptions->getOutputLike());
            $query->bindValue($i++, $like);
        }

        $query->bindValue($i++, $HostNotificationQueryOptions->getLimit(), \PDO::PARAM_INT);
        $query->bindValue($i++, $HostNotificationQueryOptions->getOffset(), \PDO::PARAM_INT);

        return $this->Backend->fetchAll($query);
    }

}
